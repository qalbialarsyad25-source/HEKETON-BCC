package dto

import (
	"time"

	"main/internal/modules/announcement/domain"

	"github.com/google/uuid"
)

type AnnouncementResponse struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	SenderID  uuid.UUID `json:"senderId"`
	EventID   uuid.UUID `json:"eventId"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func DomainToResponse(a domain.Announcement) *AnnouncementResponse {
	return &AnnouncementResponse{
		ID:        a.ID,
		Title:     a.Title,
		Body:      a.Body,
		SenderID:  a.SenderId,
		EventID:   a.EventId,
		CreatedAt: a.CreatedAt,
		UpdatedAt: a.UpdateAt,
	}
}

func DomainToResponseBatch(announcement []domain.Announcement) *[]AnnouncementResponse {
	response := make([]AnnouncementResponse, 0, len(announcement))

	for _, a := range announcement {
		response = append(response, *DomainToResponse(a))
	}

	return &response
}
