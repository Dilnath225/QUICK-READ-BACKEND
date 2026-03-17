package config

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger
var sugar *zap.SugaredLogger

func InitLogger() {
	env := os.Getenv("APP_ENV")

	var cfg zap.Config
	if env == "production" {
		cfg = zap.NewProductionConfig()

	} else {
		cfg = zap.NewDevelopmentConfig()
	}

	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var err error
	Logger, err = cfg.Build()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	sugar = Logger.Sugar()
}
