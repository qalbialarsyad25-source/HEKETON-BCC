package infrastructure

import (
	"context"
	"fmt"
	"io"

	"main/internal/config"
	"main/internal/infrastructure/database"
	"main/internal/infrastructure/hasher"
	"main/internal/infrastructure/logger"
	"main/internal/infrastructure/token"
	"main/internal/shared/ports"
	portlogger "main/internal/shared/ports/logger"

	"gorm.io/gorm"
)

func Initialize(ctx context.Context, cfg *config.Config) (*ports.Adapters, *gorm.DB, error) {
	var opened []func() error
	fail := func(step string, err error) (*ports.Adapters, *gorm.DB, error) {
		for i := len(opened) - 1; i >= 0; i-- {
			_ = opened[i]()
		}
		return nil, nil, fmt.Errorf("init %s: %w", step, err)
	}

	appLogger, err := logger.NewLogger(cfg.Logger)
	if err != nil {
		return fail("logger", err)
	}
	opened = append(opened, func() error { return CloseLogger(appLogger) })

	db, err := database.Load(ctx, cfg.Database)
	if err != nil {
		return fail("database", err)
	}
	opened = append(opened, func() error { return database.Close(db) })

	tokenService := token.NewTokenService(cfg.JWT)
	bcryptHasher := hasher.NewBcryptHasher(cfg.Bcrypt)

	return &ports.Adapters{
		Logger:       appLogger,
		TokenService: tokenService,
		Hasher:       bcryptHasher,
	}, db, nil
}

func Shutdown(adapters *ports.Adapters, db *gorm.DB) error {
	if err := database.Close(db); err != nil {
		return fmt.Errorf("close database: %w", err)
	}
	return nil
}

func CloseLogger(l portlogger.Logger) error {
	if closer, ok := l.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
