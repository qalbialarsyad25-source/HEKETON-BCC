package ports

import (
	"time"

	"github.com/google/uuid"
)

type JWTPayload struct {
	JTI    uuid.UUID
	UserID uuid.UUID
	Name   string
	Role   string
}

// JWT is a signed token. Payload.JTI is the token's own id; ExpiresAt is when it stops parsing.
type JWT struct {
	Payload   JWTPayload
	Value     string
	ExpiresAt time.Time
}

// TokenService signs and verifies JWTs. Generate* ignore data.JTI and give every token a fresh
// one. Parse* return apperrors.ErrTokenExpired for an expired token and apperrors.ErrInvalidToken
// for any other failure, including a token of the other kind (both 401).
type TokenService interface {
	GenerateAccessToken(data *JWTPayload) (*JWT, error)
	GenerateRefreshToken(data *JWTPayload) (*JWT, error)
	ParseAccessToken(tokenString string) (*JWTPayload, error)
	ParseRefreshToken(tokenString string) (*JWTPayload, error)
	AccessTTL() time.Duration
	RefreshTTL() time.Duration
}
