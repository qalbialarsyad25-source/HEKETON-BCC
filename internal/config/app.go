package config

import (
	"fmt"
	"strings"
)

// APP_MODE values.
const (
	ModeDev        = "DEV"
	ModeTest       = "TEST"
	ModeProduction = "PRODUCTION"
)

type AppConfig struct {
	Mode string
	Name string
}

func LoadAppConfig() (*AppConfig, error) {
	mode := strings.ToUpper(strings.TrimSpace(GetEnv("APP_MODE", ModeDev)))
	name := GetEnv("APP_NAME", "go-starter")

	switch mode {
	case ModeDev, ModeTest, ModeProduction:
	default:
		return nil, fmt.Errorf("APP_MODE: %q must be one of %s, %s, %s", mode, ModeDev, ModeTest, ModeProduction)
	}

	return &AppConfig{Mode: mode, Name: name}, nil
}

func (a *AppConfig) IsProduction() bool {
	return a.Mode == ModeProduction
}
