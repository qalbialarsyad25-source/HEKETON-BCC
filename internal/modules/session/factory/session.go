package factory

import (
	"time"

	"main/internal/modules/session/domain"

	"github.com/google/uuid"
)

// NewSession builds a session for unit tests.
func NewSession(id uuid.UUID) *domain.Session {
	return &domain.Session{
		ID:        id,
		EventID:   uuid.New(),
		Title:     "session-test",
		Time:      time.Now().Add(24 * time.Hour),
		Speaker:   "speaker-test",
		Location:  "location-test",
		Status:    domain.StatusUpcoming,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}
