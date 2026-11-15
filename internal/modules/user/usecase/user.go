package usecase

import (
	"context"
	"strings"

	"main/internal/modules/user/domain"
	shared "main/internal/shared/pagination"

	"github.com/google/uuid"
)

type UserRepositoryI interface {
	GetUser(ctx context.Context, page shared.Pagination) ([]domain.User, error)
	GetUserById(ctx context.Context, id uuid.UUID) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error
}

type UserUseCase struct {
	repo UserRepositoryI
}

func NewUserUseCase(repo UserRepositoryI) UserUseCase {
	return UserUseCase{repo: repo}
}

type UpdateUserInput struct {
	Name  string
	Email string
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (uc UserUseCase) GetUser(ctx context.Context, page shared.Pagination) ([]domain.User, error) {
	return uc.repo.GetUser(ctx, page)
}

func (uc UserUseCase) GetUserById(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return uc.repo.GetUserById(ctx, id)
}

func (uc UserUseCase) UpdateUser(ctx context.Context, id uuid.UUID, input UpdateUserInput) (*domain.User, error) {
	user, err := uc.repo.GetUserById(ctx, id)
	if err != nil {
		return nil, err
	}

	user.Name = input.Name
	user.Email = normalizeEmail(input.Email)

	if err := uc.repo.UpdateUser(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}


func (uc UserUseCase) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return uc.repo.DeleteUser(ctx, id)
}