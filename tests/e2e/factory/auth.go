package factory

import (
	"encoding/json"
	"net/http"
	"testing"

	"main/internal/modules/auth/controller"
	"main/internal/modules/auth/dto"
	"main/internal/modules/user/domain"
	"main/internal/shared/ports/token"
	"main/tests/e2e/helper"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// AccessToken signs an access token for user directly, without going through login.
func AccessToken(t *testing.T, tokens token.TokenService, user *domain.User) string {
	t.Helper()

	jwt, err := tokens.GenerateAccessToken(&token.JWTPayload{
		JTI:    uuid.New(),
		UserID: user.ID,
		Name:   user.Name,
	})
	require.NoError(t, err)

	return jwt.Value
}

// Login signs in through POST /api/v1/auth/login and returns the access token and the refresh cookie.
func Login(t *testing.T, router http.Handler, email, password string) (accessToken string, refreshCookie *http.Cookie) {
	t.Helper()

	w := helper.Do(t, router, http.MethodPost, "/api/v1/auth/login", dto.LoginRequest{Email: email, Password: password}, "")
	require.Equal(t, http.StatusOK, w.Code, w.Body.String())

	var body helper.SuccessResponse[dto.SessionResponse]
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	refreshCookie = helper.Cookie(w, controller.RefreshCookieName)
	require.NotNil(t, refreshCookie, "login sets the refresh cookie")

	return body.Data.AccessToken, refreshCookie
}
