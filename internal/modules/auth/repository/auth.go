package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"main/internal/infrastructure/database"
	"main/internal/modules/auth/domain"
	"main/internal/modules/auth/repository/mapper"
	"main/internal/modules/auth/usecase"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) usecase.AuthRepositoryI {
	return &AuthRepository{db: db}
}

func (repo AuthRepository) CreateRefreshToken(ctx context.Context, rec *domain.RefreshTokenRecord) error {
	if err := gorm.G[database.RefreshToken](repo.db).Create(ctx, mapper.ToDatabaseRefreshToken(rec)); err != nil {
		return fmt.Errorf("auth repository: create refresh token: %w", err)
	}

	return nil
}

func (repo AuthRepository) FindRefreshToken(ctx context.Context, id uuid.UUID) (*domain.RefreshTokenRecord, error) {
	row, err := gorm.G[database.RefreshToken](repo.db).
		Where("id = ?", id).
		First(ctx)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("auth repository: find refresh token: %w", err)
	}

	return mapper.ToDomainRefreshToken(&row), nil
}

// RevokeRefreshToken is one UPDATE with "revoked_at IS NULL" in the WHERE, so when two
// requests refresh with the same token at once, only one of them gets true.
func (repo AuthRepository) RevokeRefreshToken(ctx context.Context, id uuid.UUID, at time.Time) (bool, error) {
	updatedRows, err := gorm.G[database.RefreshToken](repo.db).
		Where("id = ? AND revoked_at IS NULL", id).
		Update(ctx, "revoked_at", at)

	if err != nil {
		return false, fmt.Errorf("auth repository: revoke refresh token: %w", err)
	}

	return updatedRows == 1, nil
}

func (repo AuthRepository) RevokeRefreshTokensOfUser(ctx context.Context, userID uuid.UUID, at time.Time) (int64, error) {
	updatedRows, err := gorm.G[database.RefreshToken](repo.db).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update(ctx, "revoked_at", at)

	if err != nil {
		return 0, fmt.Errorf("auth repository: revoke refresh tokens of user: %w", err)
	}

	return int64(updatedRows), nil
}
