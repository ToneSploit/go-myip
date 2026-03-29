package shared

import (
	"fmt"
	"main/logger"
	"os"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func SetupEnv(path string) {
	logger.Info("Initializing application settings",
		zap.String("path", path),
		zap.String("working_dir", mustGetwd()))

	// Tell viper the path/location of your env file. If it is root just add "."
	viper.AddConfigPath(path)

	// Tell viper the name of your file
	viper.SetConfigName(".env")

	// Tell viper the type of your file
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		logger.Error("Config error details",
			zap.String("error_type", fmt.Sprintf("%T", err)),
			zap.String("error", err.Error()))
		// On Railway, .env file might not exist - that's okay
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logger.Info("No .env file found, using environment variables only")
		} else {
			logger.Fatal("Error reading configuration", zap.Error(err))
		}
	} else {
		logger.Info(
			"Configuration loaded successfully", zap.String("config_file", viper.ConfigFileUsed()),
		)
	}
}

func mustGetwd() string {
	dir, err := os.Getwd()
	if err != nil {
		return "error getting working dir"
	}
	return dir
}

// Rate limit functions

func GetRateLimitConfig() (int, int, time.Duration) {
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
