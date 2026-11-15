package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	apperrors "main/internal/shared/errors"
)

// userIDContextKey is the key the auth middleware stores the signed-in user's id under.
const userIDContextKey = "userID"

var errUnauthenticated = apperrors.Unauthorized("UNAUTHENTICATED", "authentication is required")

func currentUserID(c *gin.Context) (uuid.UUID, error) {
	value, ok := c.Get(userIDContextKey)
	if !ok {
		return uuid.Nil, errUnauthenticated
	}

	switch id := value.(type) {
	case uuid.UUID:
		return id, nil
	case string:
		parsed, err := uuid.Parse(id)
		if err != nil {
			return uuid.Nil, errUnauthenticated.Wrap(err)
		}
		return parsed, nil
	default:
		return uuid.Nil, errUnauthenticated
	}
}
