package domain

import (
	"time"

	"github.com/google/uuid"
)

type Announcement struct {
	ID        uuid.UUID
	Title     string
	Body      string
	SenderId  uuid.UUID
	EventId   uuid.UUID
	CreatedAt time.Time
	UpdateAt  time.Time
}
