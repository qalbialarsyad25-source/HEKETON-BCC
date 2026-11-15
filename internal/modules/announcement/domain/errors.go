package domain

import apperrors "main/internal/shared/errors"

var (
	ErrAnnouncementNotFound  = apperrors.NotFound("ANNOUNCEMENT_NOT_FOUND", "could not find requested announcement")
	ErrInvalidAnnouncementID = apperrors.BadRequest("INVALID_ANNOUNCEMENT_ID", "announcement id must be a valid UUID")
	ErrAnnouncementForbidden = apperrors.Forbidden("ANNOUNCEMENT_FORBIDDEN", "you are not allowed to manage this announcement")
	ErrEventNotFound         = apperrors.NotFound("EVENT_NOT_FOUND", "could not find the event for this announcement")
)