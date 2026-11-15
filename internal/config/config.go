package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	App      *AppConfig
	Database *DatabaseConfig
	Logger   *LoggerConfig
	JWT      *JWTConfig
	Bcrypt   *BcryptConfig
}

func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func GetEnvInt(key string, defaultValue int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue, fmt.Errorf("%s: %q is not a valid integer", key, value)
	}
	return parsed, nil
}

func GetEnvBool(key string, defaultValue bool) (bool, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue, fmt.Errorf("%s: %q is not a valid boolean", key, value)
	}
	return parsed, nil
}

func GetEnvDuration(key string, defaultValue time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return defaultValue, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue, fmt.Errorf("%s: %q is not a valid duration (e.g. 10s, 1m)", key, value)
	}
	return parsed, nil
}

func Load() (*Config, error) {
	if err := LoadEnv(); err != nil {
		return nil, err
	}

	var errs []error
	collect := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}

	app, err := LoadAppConfig()
	if err != nil {
		collect(err)
		app = &AppConfig{Mode: ModeDev}
	}

	database, err := LoadDatabaseConfig()
	collect(err)
	logger, err := LoadLoggerConfig()
	collect(err)
	jwt, err := LoadJWTConfig(app)
	collect(err)
	bcrypt, err := LoadBcryptConfig(app)
	collect(err)

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return &Config{
		App:      app,
		Database: database,
		Logger:   logger,
		JWT:      jwt,
		Bcrypt:   bcrypt,
	}, nil
}

func LoadEnv() error {
	root, found, err := findProjectRoot()
	if err != nil || !found {
		return err
	}

	err = godotenv.Load(filepath.Join(root, ".env"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("load .env: %w", err)
	}
	return nil
}

func findProjectRoot() (string, bool, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", false, err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, true, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false, nil
		}
		dir = parent
	}
}
