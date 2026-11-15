//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"

	"main/internal/config"
	"main/internal/infrastructure/database"
	"main/internal/modules/user/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func openDB(t *testing.T) *gorm.DB {
	t.Helper()
	cfg, err := config.Load()
	require.NoError(t, err)

	db, err := database.Load(context.Background(), cfg.Database)
	require.NoError(t, err)
	t.Cleanup(func() { _ = database.Close(db) })
	return db
}

func newUser(t *testing.T, db *gorm.DB) *domain.User {
	t.Helper()
	email := fmt.Sprintf("repo-%s@example.com", uuid.NewString())
	t.Cleanup(func() { db.Where("email = ?", email).Delete(&database.User{}) })

	hash := "hashed"
	return &domain.User{ID: uuid.New(), Name: "repo user", Email: email, Password: hash}
}
