package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
)

var Log *zap.Logger = zap.NewNop()

// Initialize инициализирует логгер
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return fmt.Errorf("failed to parse log level: %w", err)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = "."
	}
	// Создаем директорию для логов
	logPath := filepath.Join(homeDir, ".gophvault", "logs", "app.log")
	err = os.MkdirAll(filepath.Dir(logPath), 0700)
	if err != nil {
		return fmt.Errorf("failed to create log dir: %w", err)
	}
	errPath := filepath.Join(homeDir, ".gophvault", "logs", "err.log")
	err = os.MkdirAll(filepath.Dir(errPath), 0700)
	if err != nil {
		return fmt.Errorf("failed to create err dir: %w", err)
	}
	// Сохраняем пути к логам
	cfg.OutputPaths = []string{logPath}
	cfg.ErrorOutputPaths = []string{errPath}

	// Конфигурируем кодировщик
	cfg.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	logger, err := cfg.Build()
	if err != nil {
		return fmt.Errorf("failed to create logger: %w", err)
	}
	Log = logger
	return nil
}
