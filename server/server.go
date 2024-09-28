package server

import (
	"flag"
	"log"
	"os"

	"github.com/ayush3160/interview-bytes-backend/pkg/models"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const defaultPort = "8080"

func flagInit(){
	models.IsDebugLevel = flag.Bool("debug", false, "Enable debug mode")

	flag.Parse()
}

func setupLogger() *zap.Logger {
	logCfg := zap.NewDevelopmentConfig()

	// Customize the encoder config to put the emoji at the beginning.
	logCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	if *models.IsDebugLevel {
		logCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		logCfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		logCfg.EncoderConfig.EncodeCaller = nil
	}

	logCfg.DisableStacktrace = true
	logger, err := logCfg.Build()
	if err != nil {
		log.Panic("failed to start the logger for the CLI")
		return nil
	}
	return logger
}

func Start() {
	flagInit()
	logger := setupLogger()

	err := godotenv.Load(".env.local")

	if err != nil {
		logger.Error("Error loading .env file", zap.Error(err))
	}

	if os.Getenv("PORT") == "" {
		os.Setenv("PORT", defaultPort)
	}

	port := os.Getenv("PORT")

	logger.Info("Starting the server", zap.String("port", port))
}