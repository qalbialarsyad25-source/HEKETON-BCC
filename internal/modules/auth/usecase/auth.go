package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"main/internal/modules/auth/domain"
	userdomain "main/internal/modules/user/domain"
	userusecase "main/internal/modules/user/usecase"
	"main/internal/shared/authz"
	"main/internal/shared/ports/token"

	"github.com/google/uuid"
)

// AuthRepositoryI stores refresh tokens. Auth owns only the refresh_tokens table.
type AuthRepositoryI interface {
	CreateRefreshToken(ctx context.Context, rec *domain.RefreshTokenRecord) error
	// FindRefreshToken returns nil, nil when no token has that id.
	FindRefreshToken(ctx context.Context, id uuid.UUID) (*domain.RefreshTokenRecord, error)
	// RevokeRefreshToken sets revoked_at only if the token is still active, in one statement,
	// and retoken whether it did. Two requests with the same token can't both get true.
	RevokeRefreshToken(ctx context.Context, id uuid.UUID, at time.Time) (bool, error)
	// RevokeRefreshTokensOfUser revokes every active token of the user and returns how many.
	RevokeRefreshTokensOfUser(ctx context.Context, userID uuid.UUID, at time.Time) (int64, error)
}

// UserService is what auth needs from the user module. The auth container fills it with
// the user usecase; auth never touches the users table itself.
type UserService interface {
	Create(ctx context.Context, input userusecase.CreateUserInput) (*userdomain.User, error)
	VerifyCredentials(ctx context.Context, input userusecase.VerifyCredentialsInput) (*userdomain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*userdomain.User, error)
}

type AuthUseCase struct {
	repo   AuthRepositoryI
	users  UserService
	tokens token.TokenService
}

// NewAuthUseCase builds the usecase. oauth may be nil: the Google methods then return
// domain.ErrOAuthDisabled.
func NewAuthUseCase(repo AuthRepositoryI, users UserService, tokens token.TokenService) *AuthUseCase {
	return &AuthUseCase{repo: repo, users: users, tokens: tokens}
}

type LoginInput struct {
	Email    string
	Password string
}

func (uc AuthUseCase) Login(ctx context.Context, input LoginInput) (*domain.Session, error) {
	user, err := uc.users.VerifyCredentials(ctx, userusecase.VerifyCredentialsInput{
		Email:    input.Email,
		Password: input.Password,
	})
	if err != nil {
		return nil, err
	}

	return uc.createSession(ctx, user)
}

type RegisterInput struct {
	Name     string
	Email    string
	Password string
}

func (uc AuthUseCase) Register(ctx context.Context, input RegisterInput) (*domain.Session, error) {
	user, err := uc.users.Create(ctx, userusecase.CreateUserInput{
		Name:     input.Name,
		Email:    input.Email,
		Password: input.Password,
		Role:     authz.RoleUser,
	})
	if err != nil {
		return nil, err
	}

	return uc.createSession(ctx, user)
}

type RefreshInput struct {
	RefreshToken string
}

// Refresh swaps a refresh token for a new session (rotation). Each refresh token works once:
// the old one is revoked before the new one is handed out. A token that was already revoked
// but is presented again was probably stolen, so every session of its user is revoked.
func (uc AuthUseCase) Refresh(ctx context.Context, input RefreshInput) (*domain.Session, error) {
	payload, err := uc.tokens.ParseRefreshToken(input.RefreshToken)
	if err != nil {
		return nil, domain.ErrRefreshTokenInvalid.Wrap(err)
	}

	now := time.Now()

	rotated, err := uc.repo.RevokeRefreshToken(ctx, payload.JTI, now)
	if err != nil {
		return nil, err
	}

	if !rotated {
		record, err := uc.repo.FindRefreshToken(ctx, payload.JTI)
		if err != nil {
			return nil, err
		}

		if record == nil {
			return nil, domain.ErrRefreshTokenInvalid
		}

		if _, err := uc.repo.RevokeRefreshTokensOfUser(ctx, record.UserID, now); err != nil {
			return nil, err
		}

		return nil, domain.ErrRefreshTokenReused
	}

	user, err := uc.users.GetByID(ctx, payload.UserID)
	if errors.Is(err, userdomain.ErrUserNotFound) {
		return nil, domain.ErrRefreshTokenInvalid.Wrap(err)
	}

	if err != nil {
		return nil, err
	}

	return uc.createSession(ctx, user)
}

type LogoutInput struct {
	RefreshToken string
}

// Logout revokes the refresh token. It is idempotent: an invalid, expired or already revoked
// token is not an error, since the session is over either way. The access token stays valid
// until it expires (it is short-lived and not stored).
func (uc AuthUseCase) Logout(ctx context.Context, input LogoutInput) error {
	payload, err := uc.tokens.ParseRefreshToken(input.RefreshToken)
	if err != nil {
		return nil
	}

	_, err = uc.repo.RevokeRefreshToken(ctx, payload.JTI, time.Now())

	return err
}

func (uc AuthUseCase) createSession(ctx context.Context, user *userdomain.User) (*domain.Session, error) {
	payload := &token.JWTPayload{
		UserID: user.ID,
		Name:   user.Name,
	}

	// Each error is checked on its own, so the second can't hide the first.
	accessToken, err := uc.tokens.GenerateAccessToken(payload)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := uc.tokens.GenerateRefreshToken(payload)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	record := domain.RefreshTokenRecord{
		ID:        refreshToken.Payload.JTI,
		UserID:    user.ID,
		ExpiresAt: refreshToken.ExpiresAt,
	}

	if err := uc.repo.CreateRefreshToken(ctx, &record); err != nil {
		return nil, err
	}

	return &domain.Session{
		AccessToken:      accessToken.Value,
		RefreshToken:     refreshToken.Value,
		AccessExpiresIn:  uc.tokens.AccessTTL(),
		RefreshExpiresIn: uc.tokens.RefreshTTL(),
		User:             *user,
	}, nil
}
