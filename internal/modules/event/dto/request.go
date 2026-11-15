package dto

import "time"

type CreateEventRequest struct {
	Title       string    `json:"title" binding:"required,min=1,max=200"`
	Description string    `json:"description" binding:"required"`
	Date        time.Time `json:"date" binding:"required"`
	Location    string    `json:"location" binding:"required,min=1,max=255"`
	MaxQuota    int       `json:"max_quota" binding:"required,min=1"`
}

// UpdateEventRequest replaces the editable fields of an event. Status is optional:
// when it is missing, the current status is kept.
type UpdateEventRequest struct {
	Title       string    `json:"title" binding:"required,min=1,max=200"`
	Description string    `json:"description" binding:"required"`
	Date        time.Time `json:"date" binding:"required"`
	Location    string    `json:"location" binding:"required,min=1,max=255"`
	MaxQuota    int       `json:"max_quota" binding:"required,min=1"`
	Status      string    `json:"status" binding:"omitempty,oneof=UPCOMING ONGOING FINISHED"`
}
