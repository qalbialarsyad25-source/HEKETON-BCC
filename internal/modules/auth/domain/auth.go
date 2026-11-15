package domain

import (
	"time"

	userdomain "main/internal/modules/user/domain"

	"github.com/google/uuid"
)

// Session is what register, login and refresh hand out: a short-lived access token,
// a long-lived refresh token and the user they belong to.
type Session struct {
	AccessToken      string
	RefreshToken     string
	AccessExpiresIn  time.Duration
	RefreshExpiresIn time.Duration
	User             userdomain.User
}

// RefreshTokenRecord is a stored refresh token. ID is the token's JTI; the token itself
// is never stored. RevokedAt is nil while the token can still be used.
type RefreshTokenRecord struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	ExpiresAt time.Time
	RevokedAt *time.Time
}
