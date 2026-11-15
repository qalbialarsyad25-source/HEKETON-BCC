package auth_test

import (
	"net/http"
	"testing"

	"main/internal/modules/auth/domain"
	"main/internal/modules/auth/dto"
	"main/internal/shared/authz"
	apperrors "main/internal/shared/errors"
	"main/internal/shared/ports/token"
	"main/tests/e2e/factory"
	"main/tests/e2e/helper"
	"main/tests/e2e/suite"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefresh(t *testing.T) {
	ts := suite.New(t)
	user := factory.CreateUser(t, ts.DB, ts.Adapters.Hasher, authz.RoleUser)

	t.Run("the cookie alone rotates into a new session (no access token needed)", func(t *testing.T) {
		_, cookie := factory.Login(t, ts.Router, user.Email, factory.Password)

		w := helper.DoWithCookies(t, ts.Router, http.MethodPost, refreshPath, nil, cookie)

		body, newCookie := assertSessionResponse(t, ts, w, http.StatusOK)
		assert.NotEqual(t, cookie.Value, newCookie.Value, "a new refresh token")
		assert.Equal(t, user.ID, body.User.ID)
	})

	t.Run("the refresh token can come in the body instead", func(t *testing.T) {
		_, cookie := factory.Login(t, ts.Router, user.Email, factory.Password)

		w := helper.Do(t, ts.Router, http.MethodPost, refreshPath, dto.RefreshRequest{RefreshToken: cookie.Value}, "")

		assertSessionResponse(t, ts, w, http.StatusOK)
	})

	t.Run("reusing a rotated token revokes every session", func(t *testing.T) {
		_, old := factory.Login(t, ts.Router, user.Email, factory.Password)
		_, otherDevice := factory.Login(t, ts.Router, user.Email, factory.Password)

		w := helper.DoWithCookies(t, ts.Router, http.MethodPost, refreshPath, nil, old)
		_, newest := assertSessionResponse(t, ts, w, http.StatusOK)

		reuse := helper.DoWithCookies(t, ts.Router, http.MethodPost, refreshPath, nil, old)
		assertError(t, reuse, http.StatusUnauthorized, helper.ErrorFrom(domain.ErrRefreshTokenReused))

		for name, cookie := range map[string]*http.Cookie{"newest": newest, "other device": otherDevice} {
			w := helper.DoWithCookies(t, ts.Router, http.MethodPost, refreshPath, nil, cookie)
			assert.Equal(t, http.StatusUnauthorized, w.Code, "the %s token is revoked too", name)
		}
	})

	t.Run("no token is REFRESH_TOKEN_MISSING", func(t *testing.T) {
		w := helper.DoWithCookies(t, ts.Router, http.MethodPost, refreshPath, nil)

		assertError(t, w, http.StatusUnauthorized, helper.ErrorFrom(domain.ErrRefreshTokenMissing))
	})

	t.Run("an empty refreshToken in the body falls back to the cookie", func(t *testing.T) {
		_, cookie := factory.Login(t, ts.Router, user.Email, factory.Password)

		w := helper.DoWithCookies(t, ts.Router, http.MethodPost, refreshPath, `{"refreshToken":""}`, cookie)

		assertSessionResponse(t, ts, w, http.StatusOK)
	})

	t.Run("garbage and access tokens are REFRESH_TOKEN_INVALID", func(t *testing.T) {
		accessToken, _ := factory.Login(t, ts.Router, user.Email, factory.Password)

		for _, token := range []string{"not-a-jwt", accessToken} {
			w := helper.Do(t, ts.Router, http.MethodPost, refreshPath, dto.RefreshRequest{RefreshToken: token}, "")

			assertError(t, w, http.StatusUnauthorized, helper.ErrorFrom(domain.ErrRefreshTokenInvalid))
		}
	})

	t.Run("a deleted user's token is REFRESH_TOKEN_INVALID", func(t *testing.T) {
		ghost := factory.CreateUser(t, ts.DB, ts.Adapters.Hasher, authz.RoleUser)
		token, err := ts.Adapters.TokenService.GenerateRefreshToken(&token.JWTPayload{UserID: ghost.ID})
		require.NoError(t, err)
		// A token that was never stored, for a user that is gone.
		require.NoError(t, ts.DB.Exec("DELETE FROM users WHERE id = ?", ghost.ID).Error)

		w := helper.Do(t, ts.Router, http.MethodPost, refreshPath, dto.RefreshRequest{RefreshToken: token.Value}, "")

		assertError(t, w, http.StatusUnauthorized, helper.ErrorFrom(domain.ErrRefreshTokenInvalid))
	})

	t.Run("malformed JSON is MALFORMED_BODY", func(t *testing.T) {
		w := helper.Do(t, ts.Router, http.MethodPost, refreshPath, `{"refreshToken":`, "")

		assertError(t, w, http.StatusBadRequest, helper.ErrorFrom(apperrors.ErrMalformedBody))
	})
}
