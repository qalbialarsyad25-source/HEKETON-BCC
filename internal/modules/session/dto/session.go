package dto

import (
	"time"

	"main/internal/modules/session/domain"
	"main/internal/shared/pagination"
)

// CreateSessionRequest is the body of POST /sessions. A new session is always UPCOMING.
type CreateSessionRequest struct {
	EventID  string    `json:"event_id" binding:"required,uuid"`
	Title    string    `json:"title" binding:"required,min=1,max=200"`
	Time     time.Time `json:"time" binding:"required"`
	Speaker  string    `json:"speaker" binding:"required,min=1,max=200"`
	Location string    `json:"location" binding:"required,min=1,max=255"`
}

// UpdateSessionRequest replaces the editable fields of a session. Status is optional:
// when it is missing, the current status is kept.
type UpdateSessionRequest struct {
	Title    string    `json:"title" binding:"required,min=1,max=200"`
	Time     time.Time `json:"time" binding:"required"`
	Speaker  string    `json:"speaker" binding:"required,min=1,max=200"`
	Location string    `json:"location" binding:"required,min=1,max=255"`
	Status   string    `json:"status" binding:"omitempty,oneof=UPCOMING ONGOING FINISHED"`
}

type SessionResponse struct {
	ID       string    `json:"id"`
	EventID  string    `json:"event_id"`
	Title    string    `json:"title"`
	Time     time.Time `json:"time"`
	Speaker  string    `json:"speaker"`
	Location string    `json:"location"`
	Status   string    `json:"status"`
}

func SessionToResponse(s domain.Session) *SessionResponse {
	return &SessionResponse{
		ID:       s.ID.String(),
		EventID:  s.EventID.String(),
		Title:    s.Title,
		Time:     s.Time,
		Speaker:  s.Speaker,
		Location: s.Location,
		Status:   string(s.Status),
	}
}

// SessionListResponse is the body of GET /sessions: {"data": [...], "meta": {...}}.
type SessionListResponse struct {
	Data []SessionResponse `json:"data"`
	Meta pagination.Meta   `json:"meta"`
}

func SessionsToListResponse(result *pagination.Result[domain.Session]) *SessionListResponse {
	data := make([]SessionResponse, 0, len(result.Items))
	for _, session := range result.Items {
		data = append(data, *SessionToResponse(session))
	}

	return &SessionListResponse{Data: data, Meta: result.Meta}
}
