// Package logger provides structured logging functionality for Radarr.
package logger

import (
	"os"

	"github.com/radarr/radarr-go/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger provides structured logging for Radarr
type Logger struct {
	*zap.SugaredLogger
}

// New creates a new logger with the given configuration
func New(cfg config.LogConfig) *Logger {
	var zapConfig zap.Config

	switch cfg.Level {
	case "debug":
		zapConfig = zap.NewDevelopmentConfig()
	case "info", "":
		zapConfig = zap.NewProductionConfig()
	case "warn":
		zapConfig = zap.NewProductionConfig()
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	case "error":
		zapConfig = zap.NewProductionConfig()
		zapConfig.Level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		zapConfig = zap.NewProductionConfig()
	}

	// Configure output
	if cfg.Output != "" && cfg.Output != "stdout" {
		zapConfig.OutputPaths = []string{cfg.Output}
		zapConfig.ErrorOutputPaths = []string{cfg.Output}
	}

	// Configure encoding
	const consoleFormat = "console"
	if cfg.Format == consoleFormat {
		zapConfig.Encoding = consoleFormat
		zapConfig.EncoderConfig = zap.NewDevelopmentEncoderConfig()
	}

	logger, err := zapConfig.Build()
	if err != nil {
		// Fallback to basic logger
		logger = zap.NewNop()
	}

	return &Logger{
		SugaredLogger: logger.Sugar(),
	}
}

func (l *Logger) Fatal(msg string, keysAndValues ...interface{}) {
	l.Fatalw(msg, keysAndValues...)
	os.Exit(1)
}

func (l *Logger) Close() {
	_ = l.Sync()
}
