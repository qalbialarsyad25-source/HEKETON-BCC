package repository

import (
	"context"
	"errors"
	"fmt"
	"main/internal/infrastructure/database"
	"main/internal/modules/user/domain"
	"main/internal/modules/user/repository/mapper"
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

func (repo UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row, err := gorm.G[database.User](repo.db).
		Where("id = ?", id).
		First(ctx)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("user repository: get by id: %w", err)
	}

	return mapper.ToDomainUser(&row), nil
}

func (repo UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	row, err := gorm.G[database.User](repo.db).
		Where("email = ?", email).
		First(ctx)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("user repository: get by email: %w", err)
	}

	return mapper.ToDomainUser(&row), nil
}

func (repo UserRepository) List(ctx context.Context, paging pagination.Pagination) (*[]domain.User, int, error) {
	rows, err := gorm.G[database.User](repo.db).
		Order("created_at DESC").
		Limit(paging.Limit).
		Offset(paging.Offset).
		Find(ctx)

	if err != nil {
		return nil, 0, fmt.Errorf("user repository: list: %w", err)
	}

	users := make([]domain.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, *mapper.ToDomainUser(&row))
	}

	count, err := gorm.G[database.User](repo.db).Count(ctx, "*")
	if err != nil {
		return nil, 0, fmt.Errorf("user repository: count: %w", err)
	}

	return &users, int(count), nil
}

func (repo UserRepository) Create(ctx context.Context, data *domain.User) (*domain.User, error) {
	dbUser := mapper.ToDatabaseUser(data)

	if err := gorm.G[database.User](repo.db).Create(ctx, dbUser); err != nil {
		if isAccountTaken(err) {
			return nil, domain.ErrEmailTaken
		}

		return nil, fmt.Errorf("user repository: create: %w", err)
	}

	return mapper.ToDomainUser(dbUser), nil
}

func (repo UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := gorm.G[database.User](repo.db).Delete(ctx)

	return err
}

func (repo UserRepository) Update(ctx context.Context, data *domain.User) (*domain.User, error) {
	var result *database.User

	err := repo.db.Transaction(func(tx *gorm.DB) error {
		updatedRows, err := gorm.G[database.User](tx).
			Where("id = ?", data.ID).
			Update(ctx, "name", data.Name)
		if err != nil {
			return err
		}

		if updatedRows == 0 {
			return nil
		}

		row, err := gorm.G[database.User](tx).
			Where("id = ?", data.ID).
			First(ctx)
		if err != nil {
			return err
		}

		result = &row

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("user repository: update: %w", err)
	}

	return mapper.ToDomainUser(result), nil
}

func isAccountTaken(err error) bool {
	code, constraint, ok := database.ConstraintViolation(err)
	return ok && code == database.UniqueViolation &&
		(constraint == "idx_users_email" || constraint == "idx_users_google_sub")
}
