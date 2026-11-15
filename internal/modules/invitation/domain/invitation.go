package domain

import (
	"time"

	"github.com/google/uuid"

	apperrors "main/internal/shared/errors"
)

var (
	ErrInvitationNotFound       = apperrors.NotFound("INVITATION_NOT_FOUND", "invitation not found")
	ErrEventNotFound            = apperrors.NotFound("EVENT_NOT_FOUND", "event not found")
	ErrReceiverNotFound         = apperrors.NotFound("USER_NOT_FOUND", "user not found")
	ErrNotEventCoordinator      = apperrors.Forbidden("NOT_EVENT_COORDINATOR", "only a coordinator of this event can do that")
	ErrInvalidRole              = apperrors.BadRequest("INVALID_ROLE", "role must be COORDINATOR or STAFF")
	ErrInvalidEventID           = apperrors.BadRequest("INVALID_EVENT_ID", "event id is not a valid UUID")
	ErrInvalidReceiverID        = apperrors.BadRequest("INVALID_RECEIVER_ID", "receiver id is not a valid UUID")
	ErrCannotInviteSelf         = apperrors.BadRequest("CANNOT_INVITE_SELF", "you cannot invite yourself")
	ErrAlreadyMember            = apperrors.Conflict("ALREADY_EVENT_MEMBER", "user is already on this event's committee")
	ErrInvitationAlreadyPending = apperrors.Conflict("INVITATION_ALREADY_PENDING", "this user already has a pending invitation to this event")
	ErrInvitationNotPending     = apperrors.Conflict("INVITATION_NOT_PENDING", "invitation has already been answered or cancelled")
)

type Role string

const (
	RoleCoordinator Role = "COORDINATOR"
	RoleStaff       Role = "STAFF"
)

func (r Role) Valid() bool {
	return r == RoleCoordinator || r == RoleStaff
}

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusAccepted  Status = "ACCEPTED"
	StatusDeclined  Status = "DECLINED"
	StatusCancelled Status = "CANCELLED"
)

// Sender is the user who sent the invitation. Only what a client needs to show it.
type Sender struct {
	ID    uuid.UUID
	Email string
}

// Invitation asks ReceiverID to join the committee of EventID with the given Role.
// Accepting it makes the receiver a committee member; declining or cancelling it doesn't.
// Only a PENDING invitation can change.
type Invitation struct {
	ID         uuid.UUID
	EventID    uuid.UUID
	Sender     Sender
	ReceiverID uuid.UUID
	Role       Role
	Status     Status
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
