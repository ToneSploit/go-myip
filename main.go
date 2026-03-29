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
	// if os.Getenv("GEOIP_ENABLED") == "true" {
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
	// app.Use(middleware.Recover()) // Add this right after creating the app
	app.Use(middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			logger.Error("panic recovered", zap.Error(err), zap.ByteString("stack", stack))
			return nil
		},
	}))

	t := template.Must(template.ParseGlob("templates/*.html"))
	app.Renderer = &TemplateRenderer{templates: t}

	app.GET("/", func(c echo.Context) error {
		// ip := c.Request().Header.Get("X-Real-Ip")
		// if ip == "" {
		// 	ip = "uncertain (could not determine client IP)"
		// }

		ip := getClientIP(c.Request())

		if !isPublicIP(net.ParseIP(ip)) {
			ip = "8.8.8.8" // fallback IP of my choice
		}

		// User-Agent
		ua := c.Request().Header.Get("User-Agent")
		if ua == "" {
			ua = "uncertain (could not determine user agent)"
		}

		println(ip)

		var loc *shared.GeoLocation
		if MaxmindDB != "" {
			var err error
			loc, err = shared.LookupIP(MaxmindDB, ip)
			if err != nil {
				logger.Error("Failed to lookup IP", zap.Error(err))
				loc.City = "uncertain (failed to lookup IP)"
				loc.Country = "uncertain (failed to lookup IP)"
				loc.Continent = "uncertain (failed to lookup IP)"
				loc.ContinentCode = "uncertain (failed to lookup IP)"
				loc.CountryCode = "uncertain (failed to lookup IP)"
				println(loc.City)
			}
		}

		// Start returning different responses based on the Accept header
		accept := c.Request().Header.Get("Accept")

		if strings.Contains(accept, "application/json") {
			return c.JSON(http.StatusOK, map[string]string{
				"ip":             ip,
				"user_agent":     ua,
				"city":           loc.City,
				"country":        loc.Country,
				"continent":      loc.Continent,
				"country_code":   loc.CountryCode,
				"continent_code": loc.ContinentCode,
			})
		}

		if strings.Contains(accept, "text/plain") {
			return c.String(http.StatusOK, fmt.Sprintf("IP: %s\nUser-Agent: %s\nCity: %s\nCountry: %s\nContinent: %s\nCountry Code: %s\nContinent Code: %s", ip, ua, loc.City, loc.Country, loc.Continent, loc.CountryCode, loc.ContinentCode))
		}

		// Default response
		return c.Render(http.StatusOK, "index.html", map[string]string{
			"IP":            ip,
			"UserAgent":     ua,
			"City":          loc.City,
			"Country":       loc.Country,
			"Continent":     loc.Continent,
			"CountryCode":   loc.CountryCode,
			"ContinentCode": loc.ContinentCode,
			"Tinylytics":    viper.GetString("TINYLYTICS"),
		})
	})

	app.GET("/health", func(c echo.Context) error {
		logger.Info("/health called")
		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "ok",
			"slogan": "All aboard the railway express!",
			"date":   time.Now().Format("2006-01-02 15:04:05"),
		})
	})

	app.GET("/echo", func(c echo.Context) error {
		var sb strings.Builder
		for k, v := range c.Request().Header {
			sb.WriteString(fmt.Sprintf("%s: %s\n", k, strings.Join(v, ", ")))
		}
		return c.String(http.StatusOK, sb.String())
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
