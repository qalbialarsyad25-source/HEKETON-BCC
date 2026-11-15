package domain

import apperrors "main/internal/shared/errors"

var (
	ErrUserNotFound       = apperrors.NotFound("USER_NOT_FOUND", "could not find requested user")
	ErrEmailTaken         = apperrors.Conflict("EMAIL_TAKEN", "email is already registered")
	ErrInvalidCredentials = apperrors.Unauthorized("INVALID_CREDENTIALS", "invalid email or password")
)
