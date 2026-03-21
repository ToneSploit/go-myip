package shared

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"main/logger"
)

func SetupEnv(path string) {
	logger.Info("Initializing application settings", zap.String("path", path))

	// Tell viper the path/location of your env file. If it is root just add "."
	viper.AddConfigPath(path)

	// Tell viper the name of your file
	viper.SetConfigName(".env")

	// Tell viper the type of your file
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
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
