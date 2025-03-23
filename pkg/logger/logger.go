package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new configured logger
func New(level string, isDevelopment bool) (*zap.Logger, error) {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	var config zap.Config
	if isDevelopment {
		// Development mode - pretty console output
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		// Production mode - JSON output
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	config.Level = zap.NewAtomicLevelAt(zapLevel)
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	// Replace the global logger to make zap.L() work
	zap.ReplaceGlobals(logger)

	return logger, nil
}

// Sync flushes any buffered log entries
func Sync(logger *zap.Logger) {
	// Ignoring the error as it's common to get an error when syncing to stdout/stderr
	_ = logger.Sync()
}

// Fatal logs a fatal message and exits the program
func Fatal(msg string, fields ...zap.Field) {
	zap.L().Fatal(msg, fields...)
	os.Exit(1)
}
