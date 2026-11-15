//go:build e2e

package auth_test

import (
	"net/http"
	"testing"

	"main/internal/modules/auth/dto"
	"main/internal/shared/authz"
	"main/tests/e2e/factory"
	"main/tests/e2e/helper"
	"main/tests/e2e/suite"

	"github.com/stretchr/testify/assert"
)

func TestLogout(t *testing.T) {
	ts := suite.New(t)
	user := factory.CreateUser(t, ts.DB, ts.Adapters.Hasher, authz.RoleUser)

	t.Run("revokes the refresh token and clears the cookie", func(t *testing.T) {
		_, cookie := factory.Login(t, ts.Router, user.Email, factory.Password)

		w := helper.DoWithCookies(t, ts.Router, http.MethodPost, logoutPath, nil, cookie)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Empty(t, w.Body.String(), "a 204 has no body")
		assertCookieCleared(t, w)

		refresh := helper.DoWithCookies(t, ts.Router, http.MethodPost, refreshPath, nil, cookie)
		assert.Equal(t, http.StatusUnauthorized, refresh.Code, "the token no longer refreshes")
	})

	t.Run("the token can come in the body", func(t *testing.T) {
		_, cookie := factory.Login(t, ts.Router, user.Email, factory.Password)

		w := helper.Do(t, ts.Router, http.MethodPost, logoutPath, dto.RefreshRequest{RefreshToken: cookie.Value}, "")
		assert.Equal(t, http.StatusNoContent, w.Code)

		refresh := helper.DoWithCookies(t, ts.Router, http.MethodPost, refreshPath, nil, cookie)
		assert.Equal(t, http.StatusUnauthorized, refresh.Code)
	})

	t.Run("without a token it is still 204 and clears the cookie", func(t *testing.T) {
		w := helper.DoWithCookies(t, ts.Router, http.MethodPost, logoutPath, nil)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Empty(t, w.Body.String())
		assertCookieCleared(t, w)
	})

	t.Run("a garbage token is still 204", func(t *testing.T) {
		w := helper.Do(t, ts.Router, http.MethodPost, logoutPath, dto.RefreshRequest{RefreshToken: "not-a-jwt"}, "")

		assert.Equal(t, http.StatusNoContent, w.Code)
		assertCookieCleared(t, w)
	})

	t.Run("logging out twice is fine", func(t *testing.T) {
		_, cookie := factory.Login(t, ts.Router, user.Email, factory.Password)

		first := helper.DoWithCookies(t, ts.Router, http.MethodPost, logoutPath, nil, cookie)
		second := helper.DoWithCookies(t, ts.Router, http.MethodPost, logoutPath, nil, cookie)

		assert.Equal(t, http.StatusNoContent, first.Code)
		assert.Equal(t, http.StatusNoContent, second.Code)
	})
}
