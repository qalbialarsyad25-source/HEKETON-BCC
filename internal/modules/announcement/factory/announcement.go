package factory

import (
	"time"

	"main/internal/modules/announcement/domain"

	"github.com/google/uuid"
)

// NewAnnouncement builds an announcement for unit tests.
func NewAnnouncement(id uuid.UUID) *domain.Announcement {
	return &domain.Announcement{
		ID:        id,
		Title:     "announcement-test",
		Body:      "announcement body test",
		SenderId:  uuid.New(),
		EventId:   uuid.New(),
		CreatedAt: time.Now(),
		UpdateAt:  time.Now(),
	}
}
