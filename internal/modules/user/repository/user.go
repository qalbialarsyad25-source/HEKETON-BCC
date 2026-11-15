package repository

import (
	"context"
	"errors"
	"main/internal/modules/user/domain"
	"main/internal/modules/user/usecase"
	"main/internal/shared/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) usecase.UserRepositoryI {
	return &UserRepository{db: db}
}

func (u *UserRepository) GetUser(ctx context.Context, page pagination.Pagination) ([]domain.User, error) {
	users, err := gorm.G[domain.User](u.db).
		Order("created_at DESC").
		Limit(page.Limit).
		Offset(page.Offset).
		Find(ctx)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (u *UserRepository) GetUserById(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user, err := gorm.G[domain.User](u.db).
		Where("id = ?", id).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (u *UserRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	return u.db.WithContext(ctx).Save(user).Error
}

func (u *UserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
	rows, err := gorm.G[domain.User](u.db).
		Where("id = ?", id).
		Delete(ctx)
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}
