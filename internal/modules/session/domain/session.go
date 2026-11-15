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

func (s Status) Valid() bool {
	return s == StatusUpcoming || s == StatusOngoing || s == StatusFinished
}

// Session is one talk or activity of an event's schedule.
type Session struct {
	ID        uuid.UUID
	EventID   uuid.UUID
	Title     string
	Time      time.Time
	Speaker   string
	Location  string
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
}
