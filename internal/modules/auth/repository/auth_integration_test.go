//go:build e2e

package repository_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"main/internal/config"
	"main/internal/infrastructure/database"
	"main/internal/modules/auth/domain"
	"main/internal/modules/auth/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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

// insertUser stores a user row (refresh tokens need one) and removes it, with its tokens, when the test ends.
func insertUser(t *testing.T, db *gorm.DB) uuid.UUID {
	t.Helper()
	id := uuid.New()
	user := &database.User{ID: id, Name: "auth repo", Email: fmt.Sprintf("auth-repo-%s@example.com", id)}
	require.NoError(t, db.Create(user).Error)
	t.Cleanup(func() { db.Delete(&database.User{}, "id = ?", id) })
	return id
}

func newRecord(userID uuid.UUID) *domain.RefreshTokenRecord {
	return &domain.RefreshTokenRecord{ID: uuid.New(), UserID: userID, ExpiresAt: time.Now().Add(time.Hour).Truncate(time.Microsecond)}
}

func TestAuthRepository_CreateAndFind(t *testing.T) {
	db := openDB(t)
	repo := repository.NewAuthRepository(db)
	ctx := context.Background()
	rec := newRecord(insertUser(t, db))

	require.NoError(t, repo.CreateRefreshToken(ctx, rec))

	found, err := repo.FindRefreshToken(ctx, rec.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, rec.ID, found.ID)
	assert.Equal(t, rec.UserID, found.UserID)
	assert.True(t, rec.ExpiresAt.Equal(found.ExpiresAt), "expires_at %v, want %v", found.ExpiresAt, rec.ExpiresAt)
	assert.Nil(t, found.RevokedAt)
}

func TestAuthRepository_FindMissingIsNilNil(t *testing.T) {
	repo := repository.NewAuthRepository(openDB(t))

	found, err := repo.FindRefreshToken(context.Background(), uuid.New())

	require.NoError(t, err)
	assert.Nil(t, found)
}

func TestAuthRepository_RevokeOnlyOnce(t *testing.T) {
	db := openDB(t)
	repo := repository.NewAuthRepository(db)
	ctx := context.Background()
	rec := newRecord(insertUser(t, db))
	require.NoError(t, repo.CreateRefreshToken(ctx, rec))
	at := time.Now().Truncate(time.Microsecond)

	first, err := repo.RevokeRefreshToken(ctx, rec.ID, at)
	require.NoError(t, err)
	second, err := repo.RevokeRefreshToken(ctx, rec.ID, at.Add(time.Minute))
	require.NoError(t, err)
	missing, err := repo.RevokeRefreshToken(ctx, uuid.New(), at)
	require.NoError(t, err)

	assert.True(t, first)
	assert.False(t, second, "an already revoked token is not revoked again")
	assert.False(t, missing)
	found, err := repo.FindRefreshToken(ctx, rec.ID)
	require.NoError(t, err)
	require.NotNil(t, found.RevokedAt)
	assert.True(t, at.Equal(*found.RevokedAt), "the first revoke time is kept")
}

func TestAuthRepository_RevokeTokensOfUser(t *testing.T) {
	db := openDB(t)
	repo := repository.NewAuthRepository(db)
	ctx := context.Background()
	userID := insertUser(t, db)
	otherID := insertUser(t, db)

	active1, active2, revoked, other := newRecord(userID), newRecord(userID), newRecord(userID), newRecord(otherID)
	for _, rec := range []*domain.RefreshTokenRecord{active1, active2, revoked, other} {
		require.NoError(t, repo.CreateRefreshToken(ctx, rec))
	}
	_, err := repo.RevokeRefreshToken(ctx, revoked.ID, time.Now())
	require.NoError(t, err)

	count, err := repo.RevokeRefreshTokensOfUser(ctx, userID, time.Now())

	require.NoError(t, err)
	assert.Equal(t, int64(2), count, "only the active tokens are counted")
	for _, rec := range []*domain.RefreshTokenRecord{active1, active2} {
		found, err := repo.FindRefreshToken(ctx, rec.ID)
		require.NoError(t, err)
		assert.NotNil(t, found.RevokedAt)
	}
	found, err := repo.FindRefreshToken(ctx, other.ID)
	require.NoError(t, err)
	assert.Nil(t, found.RevokedAt, "another user's token is untouched")
}

func TestAuthRepository_TokensGoWithTheirUser(t *testing.T) {
	db := openDB(t)
	repo := repository.NewAuthRepository(db)
	ctx := context.Background()
	userID := insertUser(t, db)
	rec := newRecord(userID)
	require.NoError(t, repo.CreateRefreshToken(ctx, rec))

	require.NoError(t, db.Delete(&database.User{}, "id = ?", userID).Error)

	found, err := repo.FindRefreshToken(ctx, rec.ID)
	require.NoError(t, err)
	assert.Nil(t, found, "ON DELETE CASCADE removes the user's tokens")
}

func TestAuthRepository_UnknownUserIsAnError(t *testing.T) {
	repo := repository.NewAuthRepository(openDB(t))

	err := repo.CreateRefreshToken(context.Background(), newRecord(uuid.New()))

	assert.ErrorContains(t, err, "auth repository: create refresh token")
}
