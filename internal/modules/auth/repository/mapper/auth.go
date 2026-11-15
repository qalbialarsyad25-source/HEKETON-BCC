package mapper

import (
	"main/internal/infrastructure/database"
	"main/internal/modules/auth/domain"
)

func ToDomainRefreshToken(t *database.RefreshToken) *domain.RefreshTokenRecord {
	if t == nil {
		return nil
	}

	return &domain.RefreshTokenRecord{
		ID:        t.ID,
		UserID:    t.UserID,
		ExpiresAt: t.ExpiresAt,
		RevokedAt: t.RevokedAt,
	}
}

func ToDatabaseRefreshToken(r *domain.RefreshTokenRecord) *database.RefreshToken {
	if r == nil {
		return nil
	}

	return &database.RefreshToken{
		ID:        r.ID,
		UserID:    r.UserID,
		ExpiresAt: r.ExpiresAt,
		RevokedAt: r.RevokedAt,
	}
}
