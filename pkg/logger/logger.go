package logger

import (
	"go.uber.org/zap"
)

const logFilePath = "logfile"

func InitLogger() {
	config := zap.Config{
		Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"stdout", logFilePath},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	zap.ReplaceGlobals(logger)

	zap.L().Info("Custom logger initialized")
}
