//go:build integration

package repository

import (
	"context"
	"fmt"
	"testing"

	"main/internal/config"
	"main/internal/infrastructure/database"
	"main/internal/modules/announcement/domain"

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

func newAnnouncement(t *testing.T, db *gorm.DB) *domain.Announcement {
	t.Helper()

	id := uuid.New()

	t.Cleanup(func() {
		db.Where("id = ?", id).Delete(&database.Announcement{})
	})

	return &domain.Announcement{
		ID:       id,
		Title:    fmt.Sprintf("announcement-%s", uuid.NewString()),
		Body:     "announcement integration test",
		SenderId: uuid.New(),
		EventId:  uuid.New(),
	}
}