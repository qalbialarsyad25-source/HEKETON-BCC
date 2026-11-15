package factory

import (
	"time"

	"main/internal/modules/user/domain"

	"github.com/google/uuid"
)

// NewUser builds a user for unit tests.
func NewUser(id uuid.UUID) *domain.User {
	hash := "hashed-password"
	return &domain.User{
		ID:        id,
		Name:      "user-test",
		Email:     "user-test@example.com",
		Password:  hash,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
