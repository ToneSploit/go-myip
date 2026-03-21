package main

import (
	"fmt"
	"main/logger"
	"main/shared"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func main() {
	logger.Info("Starting application", zap.String("version", "1.0.0"))
	shared.SetupEnv(".")
	logger.Info("Application initialized successfully")

	app := echo.New()
	// app.IPExtractor = echo.ExtractIPFromXFFHeader()
	app.IPExtractor = echo.ExtractIPFromRealIPHeader()

	app.GET("/", func(c echo.Context) error {
		logger.Info("Root endpoint called", zap.String("client_ip", c.RealIP()))
		// ip := c.RealIP()

		return c.JSON(http.StatusOK, map[string]interface{}{
			"status": "ok",
			"slogan": "Wow, so this is what it's like to be on the internet!",
			"date":   time.Now().Format("2006-01-02 15:04:05"),
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
