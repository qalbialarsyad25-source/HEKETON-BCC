package dto

import (
	"time"

	"main/internal/modules/user/domain"
	"github.com/google/uuid"
)

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updateAt"`
}

func DomainToResponse(u domain.User) *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Password:  u.Password,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func DomainToResponseBatch(users []domain.User) *[]UserResponse {
	response := make([]UserResponse, 0, len(users))

	for _, u := range users{
		response = append(response, *DomainToResponse(u))
	}

	return &response
}
