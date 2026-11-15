package config

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

type ServerConfig struct {
	Port            string
	ShutdownTimeout time.Duration
}

func LoadServerConfig() (*ServerConfig, error) {
	serverPort := GetEnv("APP_PORT", "8080")
	shutdownTimeout, timeoutErr := GetEnvDuration("SHUTDOWN_TIMEOUT", 10*time.Second)

	errs := []error{timeoutErr}
	if port, err := strconv.Atoi(serverPort); err != nil || port < 1 || port > 65535 {
		errs = append(errs, fmt.Errorf("APP_PORT: %q must be a port number between 1 and 65535", serverPort))
	}

	if timeoutErr == nil && shutdownTimeout <= 0 {
		errs = append(errs, fmt.Errorf("SHUTDOWN_TIMEOUT: must be greater than 0, got %s", shutdownTimeout))
	}

	return &ServerConfig{Port: serverPort, ShutdownTimeout: shutdownTimeout}, errors.Join(errs...)
}
