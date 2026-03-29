package main

import (
	"fmt"
	"html/template"
	"io"
	"main/logger"
	"main/shared"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

type TemplateRenderer struct {
	templates *template.Template
}

func (r *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return r.templates.ExecuteTemplate(w, name, data)
}

func main() {
	logger.Info("Starting application", zap.String("version", "1.0.0"))
	shared.SetupEnv(".")
	logger.Info("Application initialized successfully")

	// When on railway, the public domain is available in the environment variable RAILWAY_PUBLIC_DOMAIN.
	domain := os.Getenv("RAILWAY_PUBLIC_DOMAIN")
	logger.Info("Running on domain", zap.String("domain", domain))

	var MaxmindDB string
	if viper.GetString("GEOIP_ENABLED") == "true" {
		logger.Info("GeoIP lookups are enabled")
		// Download the GeoLite2 database at startup and log the result.
		var err error
		MaxmindDB, err = shared.DownloadGeoLiteDB()
		if err != nil {
			logger.Fatal("Failed to download GeoLite2 database", zap.Error(err))
		}
		logger.Info("GeoLite2 database downloaded successfully", zap.String("file", MaxmindDB))
	} else {
		logger.Info("GeoIP lookups are disabled")
	}

	app := echo.New()

	// Add middleware for logging and recovery
	app.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			logger.Error("panic recovered", zap.Error(err), zap.ByteString("stack", stack))
			return nil
		},
	}))

	limiter_rate, limiter_burst, limiter_expires := getRateLimitConfig()

	// Rate limiter — applied per route, not globally (excludes /health)
	rateLimiter := middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(limiter_rate / 60.0), // X requests per minute
				Burst:     limiter_burst,
				ExpiresIn: limiter_expires,
			},
		),
		IdentifierExtractor: func(c echo.Context) (string, error) {
			return getClientIP(c.Request()), nil
		},
		DenyHandler: func(c echo.Context, identifier string, err error) error {
			accept := c.Request().Header.Get("Accept")
			if strings.Contains(accept, "application/json") {
				return c.JSON(http.StatusTooManyRequests, map[string]string{
					"error": "rate limit exceeded, please slow down",
				})
			}
			if strings.Contains(accept, "text/plain") {
				return c.String(http.StatusTooManyRequests, "Rate limit exceeded, please slow down.\n")
			}
			return c.HTML(http.StatusTooManyRequests, "<p>Rate limit exceeded — please slow down.</p>")
		},
	})
	// app.Use(rateLimiter) // ---> Not the same rate limiter for all

	t := template.Must(template.ParseGlob("templates/*.html"))
	app.Renderer = &TemplateRenderer{templates: t}

	app.GET("/", func(c echo.Context) error {
		ip := getClientIP(c.Request())
		publicIP := isPublicIP(net.ParseIP(ip))

		// User-Agent
		ua := c.Request().Header.Get("User-Agent")
		if ua == "" {
			ua = "uncertain (could not determine user agent)"
		}

		var loc *shared.GeoLocation
		if publicIP && MaxmindDB != "" {
			var err error
			loc, err = shared.LookupIP(MaxmindDB, ip)
			if err != nil {
				logger.Error("Failed to lookup IP", zap.Error(err))
			}
		}

		// Ensure loc is never nil
		if loc == nil {
			loc = &shared.GeoLocation{}
		}

		accept := c.Request().Header.Get("Accept")

		if strings.Contains(accept, "application/json") {
			body := map[string]interface{}{
				"ip":         ip,
				"ip_public":  publicIP,
				"user_agent": ua,
			}
			if !publicIP {
				body["note"] = "Private or unroutable IP — geolocation unavailable"
			} else {
				body["city"] = loc.City
				body["country"] = loc.Country
				body["continent"] = loc.Continent
				body["country_code"] = loc.CountryCode
				body["continent_code"] = loc.ContinentCode
			}
			return c.JSON(http.StatusOK, body)
		}

		if strings.Contains(accept, "text/plain") {
			if !publicIP {
				return c.String(http.StatusOK, fmt.Sprintf(
					"IP: %s\nNote: Private or unroutable IP — geolocation unavailable\nUser-Agent: %s",
					ip, ua,
				))
			}
			return c.String(http.StatusOK, fmt.Sprintf(
				"IP: %s\nUser-Agent: %s\nCity: %s\nCountry: %s\nContinent: %s\nCountry Code: %s\nContinent Code: %s",
				ip, ua, loc.City, loc.Country, loc.Continent, loc.CountryCode, loc.ContinentCode,
			))
		}

		// Default: HTML
		return c.Render(http.StatusOK, "index.html", map[string]interface{}{
			"IP":            ip,
			"IsPublicIP":    publicIP,
			"UserAgent":     ua,
			"City":          loc.City,
			"Country":       loc.Country,
			"Continent":     loc.Continent,
			"CountryCode":   loc.CountryCode,
			"ContinentCode": loc.ContinentCode,
			"Tinylytics":    viper.GetString("TINYLYTICS"),
		})

	}, rateLimiter)

	app.GET("/health", func(c echo.Context) error {
		logger.Info("/health called")
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "ok",
			"slogan": "All aboard the railway express!",
			"date":   time.Now().Format("2006-01-02 15:04:05"),
		})
	})

	logger.Info("Starting server on port 8080")
	if err := app.Start(":8080"); err != nil {
		logger.Fatal("Server failed to start", zap.Error(err))
	}
}

func isPublicIP(ip net.IP) bool {
	if ip == nil {
		return false
	}
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return false
	}
	return true
}

// Get the client IP address
func getClientIP(r *http.Request) string {
	// First check X-Real-Ip header
	ip := r.Header.Get("X-Real-Ip")
	if ip != "" {
		return ip
	}

	// If X-Real-Ip is not set, check X-Forwarded-For
	ip = r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// X-Forwarded-For may contain multiple IPs, take the first one
		ips := strings.Split(ip, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// If no proxy headers are available, get the direct IP
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		// If there's an error splitting the address, just return it as is
		return r.RemoteAddr
	}

	return ip
}

func getRateLimitConfig() (int, int, time.Duration) {
	// Get rate with fallback to default
	rate := viper.GetInt("APP_RATE_LIMIT_REQUESTS_PER_MINUTE")
	if rate <= 0 {
		rate = 20 // Default value
	}

	// Get burst with fallback to default
	burst := viper.GetInt("APP_RATE_LIMIT_BURST")
	if burst <= 0 {
		burst = 5 // Default value
	}

	// Get expiration in seconds with fallback
	expiresSeconds := viper.GetInt("APP_RATE_LIMIT_EXPIRES_SECONDS")
	if expiresSeconds <= 0 {
		expiresSeconds = 360 // Default 1 hour
	}

	return rate, burst, time.Duration(expiresSeconds) * time.Second
}
