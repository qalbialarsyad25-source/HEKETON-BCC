package config

import (
	"errors"
	"fmt"
)

type JWTConfig struct {
	AccessSecret        string
	RefreshSecret       string
	AccessExpiryMinutes int
	RefreshExpiryDays   int
}

func LoadJWTConfig(app *AppConfig) (*JWTConfig, error) {
	accessSecret := GetEnv("JWT_ACCESS_SECRET", DefaultJWTSecret)
	refreshSecret := GetEnv("JWT_REFRESH_SECRET", DefaultJWTSecret)
	accessExpiry, accessErr := GetEnvInt("JWT_ACCESS_EXPIRY_MINUTES", 60)
	refreshExpiry, refreshErr := GetEnvInt("JWT_REFRESH_EXPIRY_DAYS", 7)

	errs := []error{accessErr, refreshErr}
	if accessErr == nil && accessExpiry <= 0 {
		errs = append(errs, fmt.Errorf("JWT_ACCESS_EXPIRY_MINUTES: must be greater than 0, got %d", accessExpiry))
	}
	if refreshErr == nil && refreshExpiry <= 0 {
		errs = append(errs, fmt.Errorf("JWT_REFRESH_EXPIRY_DAYS: must be greater than 0, got %d", refreshExpiry))
	}
	if app.IsProduction() {
		errs = append(errs, checkProductionSecret("JWT_ACCESS_SECRET", accessSecret), checkProductionSecret("JWT_REFRESH_SECRET", refreshSecret))
		if accessSecret == refreshSecret {
			errs = append(errs, errors.New("JWT_ACCESS_SECRET and JWT_REFRESH_SECRET must differ in production"))
		}
	}

	return &JWTConfig{
		AccessSecret:        accessSecret,
		RefreshSecret:       refreshSecret,
		AccessExpiryMinutes: accessExpiry,
		RefreshExpiryDays:   refreshExpiry,
	}, errors.Join(errs...)
}

// DefaultJWTSecret is the development fallback. Production refuses to start with it.
const DefaultJWTSecret = "your-secret-key"

// minProductionSecretBytes is the shortest secret production accepts (256 bits for HS256).
const minProductionSecretBytes = 32

// checkProductionSecret stops anyone from forging tokens with a known or guessable secret.
func checkProductionSecret(key, secret string) error {
	if secret == DefaultJWTSecret {
		return fmt.Errorf("%s: the default secret is not allowed in production", key)
	}
	if len(secret) < minProductionSecretBytes {
		return fmt.Errorf("%s: must be at least %d bytes in production, got %d", key, minProductionSecretBytes, len(secret))
	}
	return nil
}
