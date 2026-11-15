package dto

// CreateInvitationRequest is the body of POST /invitations.
type CreateInvitationRequest struct {
	EventID    string `json:"event_id" binding:"required,uuid"`
	ReceiverID string `json:"receiver_id" binding:"required,uuid"`
	Role       string `json:"role" binding:"required,oneof=COORDINATOR STAFF"`
}

// InviteUserRequest is the body of POST /users/:id/invitations. The receiver comes from the path.
type InviteUserRequest struct {
	EventID string `json:"event_id" binding:"required,uuid"`
	Role    string `json:"role" binding:"required,oneof=COORDINATOR STAFF"`
}

// UpdateInvitationRequest is the body of PUT /invitation/:id.
type UpdateInvitationRequest struct {
	Role string `json:"role" binding:"required,oneof=COORDINATOR STAFF"`
}
