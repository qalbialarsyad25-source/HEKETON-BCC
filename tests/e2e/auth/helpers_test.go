package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"main/internal/modules/auth/controller"
	"main/internal/modules/auth/dto"
	userdto "main/internal/modules/user/dto"
	"main/tests/e2e/helper"
	"main/tests/e2e/suite"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	registerPath = "/api/v1/auth/register"
	loginPath    = "/api/v1/auth/login"
	refreshPath  = "/api/v1/auth/refresh"
	logoutPath   = "/api/v1/auth/logout"
	cookiePath   = "/api/v1/auth"
)

// assertSessionResponse checks a register/login/refresh answer: the body, the refresh cookie,
// and that the access token works. It returns the body and the cookie.
func assertSessionResponse(t *testing.T, ts *suite.Suite, w *httptest.ResponseRecorder, wantStatus int) (dto.SessionResponse, *http.Cookie) {
	t.Helper()
	require.Equal(t, wantStatus, w.Code, w.Body.String())

	body := helper.ParseJSON(t, w, helper.SuccessResponse[dto.SessionResponse]{}).(helper.SuccessResponse[dto.SessionResponse]).Data
	assert.NotEmpty(t, body.AccessToken)
	assert.Equal(t, "Bearer", body.TokenType)
	assert.Equal(t, int(ts.Adapters.TokenService.AccessTTL().Seconds()), body.ExpiresIn)

	cookie := helper.Cookie(w, controller.RefreshCookieName)
	require.NotNil(t, cookie, "the refresh cookie is set")
	assert.NotEmpty(t, cookie.Value)
	assert.Equal(t, body.RefreshToken, cookie.Value, "the body and the cookie carry the same refresh token")
	assert.Equal(t, cookiePath, cookie.Path)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	assert.Equal(t, int(ts.Adapters.TokenService.RefreshTTL().Seconds()), cookie.MaxAge, "the cookie lives as long as the refresh token")
	assert.Equal(t, ts.Config.HTTP.CookieSecure, cookie.Secure)

	// The access token works on a protected route.
	me := helper.Do(t, ts.Router, http.MethodGet, "/api/v1/users/me", nil, body.AccessToken)
	require.Equal(t, http.StatusOK, me.Code, me.Body.String())
	meBody := helper.ParseJSON(t, me, helper.SuccessResponse[userdto.UserResponse]{}).(helper.SuccessResponse[userdto.UserResponse])
	assert.Equal(t, body.User.ID, meBody.Data.ID)

	return body, cookie
}

// assertCookieCleared checks that the response deletes the refresh cookie on the same path.
func assertCookieCleared(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	cookie := helper.Cookie(w, controller.RefreshCookieName)
	require.NotNil(t, cookie, "the response clears the refresh cookie")
	assert.Empty(t, cookie.Value)
	assert.Negative(t, cookie.MaxAge, "Max-Age=0 deletes the cookie")
	assert.Equal(t, cookiePath, cookie.Path)
	assert.True(t, cookie.HttpOnly)
}

func assertError(t *testing.T, w *httptest.ResponseRecorder, wantStatus int, want helper.ErrorResponse) {
	t.Helper()
	assert.Equal(t, wantStatus, w.Code)
	assert.Equal(t, want, helper.ParseJSON(t, w, helper.ErrorResponse{}))
}
