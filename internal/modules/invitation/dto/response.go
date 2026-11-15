package dto

import (
	"main/internal/modules/invitation/domain"
	"main/internal/shared/pagination"
)

type SenderResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type InvitationResponse struct {
	ID         string         `json:"id"`
	EventID    string         `json:"event_id"`
	ReceiverID string         `json:"receiver_id"`
	Sender     SenderResponse `json:"sender"`
	Role       string         `json:"role"`
	Status     string         `json:"status"`
}

func InvitationToResponse(i domain.Invitation) *InvitationResponse {
	return &InvitationResponse{
		ID:         i.ID.String(),
		EventID:    i.EventID.String(),
		ReceiverID: i.ReceiverID.String(),
		Sender: SenderResponse{
			ID:    i.Sender.ID.String(),
			Email: i.Sender.Email,
		},
		Role:   string(i.Role),
		Status: string(i.Status),
	}
}

// InvitationListResponse is the body of the paginated invitation lists: {"data": [...], "meta": {...}}.
type InvitationListResponse struct {
	Data []InvitationResponse `json:"data"`
	Meta pagination.Meta      `json:"meta"`
}

func InvitationsToListResponse(result *pagination.Result[domain.Invitation]) *InvitationListResponse {
	data := make([]InvitationResponse, 0, len(result.Items))
	for _, invitation := range result.Items {
		data = append(data, *InvitationToResponse(invitation))
	}

	return &InvitationListResponse{Data: data, Meta: result.Meta}
}
