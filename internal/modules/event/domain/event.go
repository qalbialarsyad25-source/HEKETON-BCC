package domain

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusUpcoming Status = "UPCOMING"
	StatusOngoing  Status = "ONGOING"
	StatusFinished Status = "FINISHED"
)

// Event is something people can register for. AvailableQuota is MaxQuota minus the
// participants that are registered (cancelled registrations don't count).
type Event struct {
	ID             uuid.UUID
	Title          string
	Description    string
	Date           time.Time
	Location       string
	AvailableQuota int
	MaxQuota       int
	Status         Status
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
