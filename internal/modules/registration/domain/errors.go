package domain

import "errors"

var (
	ErrRegistrationNotFound   = errors.New("registration not found")
	ErrEventNotFound          = errors.New("event not found")
	ErrEventNotOpen           = errors.New("event is not open for registration")
	ErrEventQuotaFull         = errors.New("event registration quota is full")
	ErrAlreadyRegistered      = errors.New("user is already registered for this event")
	ErrAlreadyCancelled       = errors.New("registration is already cancelled")
	ErrAlreadyCheckedIn       = errors.New("participant is already checked in")
	ErrCannotCancelCheckedIn  = errors.New("cannot cancel a registration that is already checked in")
	ErrCannotCheckInCancelled = errors.New("cannot check in a cancelled registration")
	ErrInvalidStatus          = errors.New("invalid registration status")
	ErrForbidden              = errors.New("you are not authorized to perform this action")
)
