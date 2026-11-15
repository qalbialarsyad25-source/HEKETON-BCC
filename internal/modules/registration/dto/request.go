package dto

type UpdateRegistrationRequest struct {
	Status string `json:"status" binding:"required,oneof=REGISTERED CHECKEDIN CANCELLED"`
}
