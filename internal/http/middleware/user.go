package middleware

import (
	"main/internal/shared/authz"
	apperrors "main/internal/shared/errors"
	"main/internal/shared/ports/token"

	"github.com/gin-gonic/gin"
)

const UserKey = "user"

// GetJWTPayload returns the payload stored by Authenticate, or ErrUnauthorized when
// the route isn't behind Authenticate.
func GetJWTPayload(c *gin.Context) (*token.JWTPayload, error) {
	value, exist := c.Get(UserKey)
	if !exist {
		return nil, apperrors.ErrUnauthorized
	}

	user, ok := value.(*token.JWTPayload)

	if !ok || user == nil {
		return nil, apperrors.ErrUnauthorized
	}

	return user, nil
}

// GetActor returns who is calling, from the payload stored by Authenticate, for usecase inputs.
// It returns ErrUnauthorized when the route isn't behind Authenticate.
func GetActor(c *gin.Context) (authz.Actor, error) {
	user, err := GetJWTPayload(c)
	if err != nil {
		return authz.Actor{}, err
	}

	return authz.Actor{UserID: user.UserID, Role: user.Role}, nil
}
