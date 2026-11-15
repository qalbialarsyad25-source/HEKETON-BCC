package domain

import apperrors "main/internal/shared/errors"

var (
	ErrRefreshTokenInvalid = apperrors.Unauthorized("REFRESH_TOKEN_INVALID", "refresh token is invalid or expired")
	ErrRefreshTokenReused  = apperrors.Unauthorized("REFRESH_TOKEN_REUSED", "refresh token was already used, every session has been signed out")
	ErrRefreshTokenMissing = apperrors.Unauthorized("REFRESH_TOKEN_MISSING", "refresh token is required")

	ErrOAuthDisabled        = apperrors.NotFound("OAUTH_DISABLED", "google sign-in is not configured")
	ErrOAuthDenied          = apperrors.BadRequest("OAUTH_DENIED", "google sign-in was cancelled or refused")
	ErrOAuthStateInvalid    = apperrors.BadRequest("OAUTH_STATE_INVALID", "sign-in request is invalid or expired, start again")
	ErrOAuthCodeMissing     = apperrors.BadRequest("OAUTH_CODE_MISSING", "authorization code is missing")
	ErrOAuthExchangeFailed  = apperrors.BadGateway("OAUTH_EXCHANGE_FAILED", "could not complete sign-in with google")
	ErrOAuthEmailUnverified = apperrors.Forbidden("OAUTH_EMAIL_UNVERIFIED", "the google account's email is not verified")
)
