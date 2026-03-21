package main

import (
	"fmt"
	"html/template"
	"io"
	"main/logger"
	"main/shared"
	"net/http"
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

	app := echo.New()

	// Not working for some reason, always returns the IP of the load balancer instead of the client IP.
	// app.IPExtractor = echo.ExtractIPFromXFFHeader()
	// app.IPExtractor = echo.ExtractIPFromRealIPHeader()

	t := template.Must(template.ParseGlob("templates/*.html"))
	app.Renderer = &TemplateRenderer{templates: t}

	app.GET("/", func(c echo.Context) error {
		ip := c.Request().Header.Get("X-Real-Ip")
		if ip == "" {
			ip = "uncertain (could not determine client IP)"
		}

		// logger.Info("Root endpoint called", zap.String("client_ip", ip))

		accept := c.Request().Header.Get("Accept")
		if strings.Contains(accept, "application/json") {
			return c.JSON(http.StatusOK, map[string]string{
				"ip": ip,
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
