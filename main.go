package main

import (
	"fmt"
	"html/template"
	"io"
	"main/logger"
	"main/shared"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
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

	app := echo.New()

	t := template.Must(template.ParseGlob("templates/*.html"))
	app.Renderer = &TemplateRenderer{templates: t}

	app.GET("/", func(c echo.Context) error {
		ip := c.Request().Header.Get("X-Real-Ip")
		if ip == "" {
			ip = "uncertain (could not determine client IP)"
		}

		// User-Agent
		ua := c.Request().Header.Get("User-Agent")
		if ua == "" {
			ua = "uncertain (could not determine user agent)"
		}

		accept := c.Request().Header.Get("Accept")
		if strings.Contains(accept, "application/json") {
			return c.JSON(http.StatusOK, map[string]string{
				"ip":         ip,
				"user_agent": ua,
			})
		}

		return c.Render(http.StatusOK, "index.html", map[string]string{"IP": ip})
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
