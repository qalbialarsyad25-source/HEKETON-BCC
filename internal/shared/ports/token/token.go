package token

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

type JWT struct {
	Payload   JWTPayload
	Value     string
	ExpiresAt time.Time
}

type TokenService interface {
	GenerateAccessToken(data *JWTPayload) (*JWT, error)
	GenerateRefreshToken(data *JWTPayload) (*JWT, error)
	ParseAccessToken(tokenString string) (*JWTPayload, error)
	ParseRefreshToken(tokenString string) (*JWTPayload, error)
	AccessTTL() time.Duration
	RefreshTTL() time.Duration
}
