package config

import (
	"errors"
	"fmt"
	"strings"
)

type LoggerConfig struct {
	Level      string
	Format     string
	AddSource  bool
	FilePath   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
}

func LoadLoggerConfig() (*LoggerConfig, error) {
	level := strings.ToLower(GetEnv("LOG_LEVEL", "info"))
	format := strings.ToLower(GetEnv("LOG_FORMAT", "json"))
	addSource, addSourceErr := GetEnvBool("LOG_ADD_SOURCE", true)
	filePath := GetEnv("LOG_FILE_PATH", "logs/app.log")
	maxSize, maxSizeErr := GetEnvInt("LOG_MAX_SIZE", 100)
	maxBackups, maxBackupsErr := GetEnvInt("LOG_MAX_BACKUPS", 3)
	maxAge, maxAgeErr := GetEnvInt("LOG_MAX_AGE", 28)
	compress, compressErr := GetEnvBool("LOG_COMPRESS", true)

	errs := []error{addSourceErr, maxSizeErr, maxBackupsErr, maxAgeErr, compressErr}

	switch level {
	case "debug", "info", "warn", "error":
	default:
		errs = append(errs, fmt.Errorf("LOG_LEVEL: %q must be one of debug, info, warn, error", level))
	}

	switch format {
	case "json", "text":
	default:
		errs = append(errs, fmt.Errorf("LOG_FORMAT: %q must be json or text", format))
	}

	return &LoggerConfig{
		Level:      level,
		Format:     format,
		AddSource:  addSource,
		FilePath:   filePath,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   compress,
	}, errors.Join(errs...)
}
