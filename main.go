package main

import (
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"main/logger"
	"main/shared"
	"net/http"
	"time"
)

func main() {
	logger.Info("Starting application", zap.String("version", "1.0.0"))
	shared.SetupEnv(".")
	logger.Info("Application initialized successfully")

	app := echo.New()

	app.GET("/", func(c echo.Context) error {
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

	logger.Info("Starting server on port 8080")
	if err := app.Start(":8080"); err != nil {
		logger.Fatal("Server failed to start", zap.Error(err))
	}
}
