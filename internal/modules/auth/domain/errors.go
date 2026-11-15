package domain

import apperrors "main/internal/shared/errors"

var (
	ErrRefreshTokenInvalid = apperrors.Unauthorized("REFRESH_TOKEN_INVALID", "refresh token is invalid or expired")
	ErrRefreshTokenReused  = apperrors.Unauthorized("REFRESH_TOKEN_REUSED", "refresh token was already used, every session has been signed out")
	ErrRefreshTokenMissing = apperrors.Unauthorized("REFRESH_TOKEN_MISSING", "refresh token is required")
)
