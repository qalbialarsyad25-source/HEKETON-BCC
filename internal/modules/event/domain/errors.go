package domain

import apperrors "main/internal/shared/errors"

var (
	ErrEventNotFound           = apperrors.NotFound("EVENT_NOT_FOUND", "event not found")
	ErrNotEventCoordinator     = apperrors.Forbidden("NOT_EVENT_COORDINATOR", "only a coordinator of this event can do that")
	ErrEventDateInPast         = apperrors.BadRequest("EVENT_DATE_IN_PAST", "event date must be in the future")
	ErrQuotaBelowRegistrations = apperrors.Conflict("QUOTA_BELOW_REGISTRATIONS", "max quota is lower than the number of participants already registered")
)
