package middleware

import (
	"strings"

	apperrors "main/internal/shared/errors"
	"main/internal/shared/ports/token"

	"github.com/gin-gonic/gin"
)

// Authenticate requires "Authorization: Bearer <access token>" and stores the token's payload
// under UserKey (read it with GetJWTPayload). A missing or malformed header is UNAUTHORIZED,
// a bad token INVALID_TOKEN or TOKEN_EXPIRED; all are 401s rendered by ErrorHandler.
func Authenticate(tokens ports.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken, ok := bearerToken(c.GetHeader("Authorization"))
		if !ok {
			_ = c.Error(apperrors.ErrUnauthorized)
			c.Abort()
			return
		}

		accessTokenPayload, err := tokens.ParseAccessToken(accessToken)
		if err != nil {
			_ = c.Error(err)
			c.Abort()
			return
		}

		c.Set(UserKey, accessTokenPayload)

		c.Next()
	}
}

// bearerToken returns the token of a "Bearer <token>" header. The scheme is case-insensitive,
// and the token must be one non-empty word.
func bearerToken(header string) (string, bool) {
	scheme, token, found := strings.Cut(strings.TrimSpace(header), " ")
	if !found || !strings.EqualFold(scheme, "Bearer") {
		return "", false
	}

	token = strings.TrimSpace(token)
	if token == "" || strings.ContainsAny(token, " \t") {
		return "", false
	}

	return token, true
}
