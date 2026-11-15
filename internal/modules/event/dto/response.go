package dto

import (
	"time"

	"main/internal/modules/event/domain"
	"main/internal/shared/pagination"
)

type EventResponse struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Description    string    `json:"description"`
	Date           time.Time `json:"date"`
	Location       string    `json:"location"`
	AvailableQuota int       `json:"available_quota"`
	MaxQuota       int       `json:"max_quota"`
	Status         string    `json:"status"`
}

func EventToResponse(e domain.Event) *EventResponse {
	return &EventResponse{
		ID:             e.ID.String(),
		Title:          e.Title,
		Description:    e.Description,
		Date:           e.Date,
		Location:       e.Location,
		AvailableQuota: e.AvailableQuota,
		MaxQuota:       e.MaxQuota,
		Status:         string(e.Status),
	}
}

// EventListResponse is the body of the paginated event lists: {"data": [...], "meta": {...}}.
type EventListResponse struct {
	Data []EventResponse `json:"data"`
	Meta pagination.Meta `json:"meta"`
}

func EventsToListResponse(result *pagination.Result[domain.Event]) *EventListResponse {
	data := make([]EventResponse, 0, len(result.Items))
	for _, event := range result.Items {
		data = append(data, *EventToResponse(event))
	}

	return &EventListResponse{Data: data, Meta: result.Meta}
}
