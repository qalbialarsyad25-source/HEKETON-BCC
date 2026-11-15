package config

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const minProductionCost = 10

type BcryptConfig struct {
	Cost int
}

func LoadBcryptConfig(app *AppConfig) (*BcryptConfig, error) {
	cost, err := GetEnvInt("BCRYPT_COST", 10)
	if err != nil {
		return nil, err
	}
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return nil, fmt.Errorf("BCRYPT_COST: must be between %d and %d, got %d", bcrypt.MinCost, bcrypt.MaxCost, cost)
	}

	if app.IsProduction() && cost < minProductionCost {
		return nil, fmt.Errorf("BCRYPT_COST: must be at least %d in production, got %d", minProductionCost, cost)
	}

	return &BcryptConfig{Cost: cost}, nil
}
