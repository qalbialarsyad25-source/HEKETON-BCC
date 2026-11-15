package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"main/internal/config"
	"main/internal/shared/ports/logger"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

type SLogger struct {
	*slog.Logger
	file io.Closer
}

// NewLogger writes every log line to stdout and to a rotating file (cfg.FilePath).
func NewLogger(cfg *config.LoggerConfig) (logger.Logger, error) {
	var level slog.Level

	switch strings.ToLower(cfg.Level) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		return nil, fmt.Errorf("unknown log level %q", cfg.Level)
	}

	file := &lumberjack.Logger{
		Filename:   cfg.FilePath,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}

	writers := []io.Writer{
		os.Stdout,
		file,
	}

	writer := io.MultiWriter(writers...)

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
	}

	var handler slog.Handler
	switch strings.ToLower(cfg.Format) {
	case "json":
		handler = slog.NewJSONHandler(writer, opts)
	case "text":
		handler = slog.NewTextHandler(writer, opts)
	default:
		return nil, fmt.Errorf("unknown log format %q", cfg.Format)
	}

	// Created only once the config is known to be valid.
	if err := os.MkdirAll(filepath.Dir(cfg.FilePath), 0o755); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}

	return &SLogger{
		Logger: slog.New(handler),
		file:   file,
	}, nil
}

// Close closes the rotating log file. Call it on shutdown.
func (logger *SLogger) Close() error {
	return logger.file.Close()
}

func (logger SLogger) RequestLog(requestID string, method string, path string, statusCode int, start time.Time) {
	logger.Info(
		"request completed",
		slog.String("request_id", requestID),
		slog.String("method", method),
		slog.String("path", path),
		slog.Int("status", statusCode),
		slog.Duration("duration", time.Since(start)),
	)
}

func (logger SLogger) ErrorLog(requestID string, method string, path string, statusCode int, code string, err error) {
	level := slog.LevelWarn
	if statusCode >= 500 {
		level = slog.LevelError
	}

	logger.Log(
		context.Background(),
		level,
		"request failed",
		slog.String("request_id", requestID),
		slog.String("method", method),
		slog.String("path", path),
		slog.Int("status", statusCode),
		slog.String("code", code),
		slog.Any("error", err),
	)
}

var _ logger.Logger = (*SLogger)(nil)
