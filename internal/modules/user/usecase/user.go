package usecase

import (
	"context"
	"fmt"
	"strings"

	"main/internal/modules/user/domain"
	"main/internal/shared/authz"
	"main/internal/shared/pagination"
	shared "main/internal/shared/pagination"
	"main/internal/shared/ports/hasher"

	"github.com/google/uuid"
)

type UserRepositoryI interface {
	List(ctx context.Context, paging pagination.Pagination) (*[]domain.User, int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, data *domain.User) (*domain.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Create(ctx context.Context, data *domain.User) (*domain.User, error)
}

type UserUseCase struct {
	repo   UserRepositoryI
	hasher hasher.Hasher
}

func NewUserUseCase(repo UserRepositoryI, hasher hasher.Hasher) UserUseCase {
	return UserUseCase{repo: repo}
}

type UpdateUserInput struct {
	Name  string
	Email string
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (uc UserUseCase) List(ctx context.Context, page shared.Pagination) (*[]domain.User, int, error) {
	return uc.repo.List(ctx, page)
}

func (uc UserUseCase) GetById(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc UserUseCase) Update(ctx context.Context, id uuid.UUID, input UpdateUserInput) (*domain.User, error) {
	user, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	user.Name = input.Name
	user.Email = normalizeEmail(input.Email)

	if _, err := uc.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (uc UserUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	return uc.repo.Delete(ctx, id)
}

type CreateUserInput struct {
	Name     string
	Email    string
	Password string
	Role     string
}

// MaxPasswordBytes is the longest password bcrypt accepts.
const MaxPasswordBytes = 72

// Create stores a new user with a hashed password. Used by auth register and by the seed.
func (uc UserUseCase) Create(ctx context.Context, input CreateUserInput) (*domain.User, error) {
	if len(input.Password) > MaxPasswordBytes {
		return nil, domain.ErrPasswordTooLong
	}

	email := normalizeEmail(input.Email)

	// Only for a clean error before hashing; the unique index is the real guard
	// (two requests can both pass this check, and Create then returns ErrEmailTaken).
	existingUser, err := uc.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if existingUser != nil {
		return nil, domain.ErrEmailTaken
	}

	hashedPassword, err := uc.hasher.Hash(input.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	role := input.Role
	if role == "" {
		role = authz.RoleUser
	}

	newUser := domain.User{
		ID:       uuid.New(),
		Name:     input.Name,
		Email:    email,
		Password: hashedPassword,
	}

	createdUser, err := uc.repo.Create(ctx, &newUser)
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}

package usecase

import (
	"context"
	"errors"
	"fmt"

	"main/internal/modules/user/domain"
	"main/internal/shared/ports"
)

type VerifyCredentialsInput struct {
	Email    string
	Password string
}

func (uc UserUseCase) VerifyCredentials(ctx context.Context, input VerifyCredentialsInput) (*domain.User, error) {
	user, err := uc.repo.GetByEmail(ctx, normalizeEmail(input.Email))
	if err != nil {
		return nil, err
	}

	// A Google-only account has no password, so no password can match it.
	if user == nil || user.PasswordHash == nil {
		return nil, domain.ErrInvalidCredentials
	}

	if err := uc.hasher.Compare(*user.PasswordHash, input.Password); err != nil {
		if errors.Is(err, ports.ErrHashMismatch) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("compare password: %w", err)
	}

	return user, nil
}
