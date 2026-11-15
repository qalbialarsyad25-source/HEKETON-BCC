package factory

import (
	"context"
	"testing"

	"main/internal/infrastructure/database"
	"main/internal/modules/user/domain"
	"main/internal/modules/user/repository/mapper"
	"main/internal/shared/ports/hasher"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// Password is the plain password of every user CreateUser inserts.
const Password = "password123"

// CreateUser inserts a user with the given role, a unique email and the password Password.
func CreateUser(t *testing.T, db *gorm.DB, hasher hasher.Hasher, role string) *domain.User {
	t.Helper()

	hashedPassword, err := hasher.Hash(Password)
	require.NoError(t, err)

	id := uuid.New()
	user := &database.User{
		ID:       id,
		Name:     role + " " + id.String()[:8],
		Email:    role + "-" + id.String()[:8] + "@example.com",
		Password: hashedPassword,
	}

	err = gorm.G[database.User](db).Create(context.Background(), user)
	require.NoError(t, err)

	return mapper.ToDomainUser(user)
}
