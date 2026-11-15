package dto

import "main/internal/modules/registration/domain"

type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type RegistrationResponse struct {
	ID      string       `json:"id"`
	Status  string       `json:"status"`
	User    UserResponse `json:"user"`
	EventID string       `json:"event_id"`
}

func ToRegistrationResponse(r *domain.Registration) RegistrationResponse {
	return RegistrationResponse{
		ID:      r.ID,
		Status:  string(r.Status),
		EventID: r.EventID,
		User: UserResponse{
			ID:    r.User.ID,
			Email: r.User.Email,
		},
	}
}

func ToRegistrationListResponse(items []domain.Registration) []RegistrationResponse {
	res := make([]RegistrationResponse, 0, len(items))
	for i := range items {
		res = append(res, ToRegistrationResponse(&items[i]))
	}
	return res
}
