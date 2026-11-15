package domain

import apperrors "main/internal/shared/errors"

var (
	ErrSessionNotFound = apperrors.NotFound("SESSION_NOT_FOUND", "session not found")
	ErrEventNotFound   = apperrors.NotFound("EVENT_NOT_FOUND", "event not found")
	ErrNotEventMember  = apperrors.Forbidden("NOT_EVENT_COMMITTEE_MEMBER", "only a committee member of this event can do that")
	ErrInvalidEventID  = apperrors.BadRequest("INVALID_EVENT_ID", "event id is not a valid UUID")
	ErrInvalidStatus   = apperrors.BadRequest("INVALID_STATUS", "status must be UPCOMING, ONGOING or FINISHED")
)
