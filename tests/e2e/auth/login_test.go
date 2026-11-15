//go:build e2e

package auth_test

import (
	"net/http"
	"strings"
	"testing"

	"main/internal/modules/auth/dto"
	userdomain "main/internal/modules/user/domain"
	"main/internal/shared/authz"
	apperrors "main/internal/shared/errors"
	"main/tests/e2e/factory"
	"main/tests/e2e/helper"
	"main/tests/e2e/suite"

	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {
	ts := suite.New(t)
	user := factory.CreateUser(t, ts.DB, ts.Adapters.Hasher, authz.RoleUser)

	t.Run("signs in with email and password", func(t *testing.T) {
		w := helper.Do(t, ts.Router, http.MethodPost, loginPath, dto.LoginRequest{Email: user.Email, Password: factory.Password}, "")

		body, _ := assertSessionResponse(t, ts, w, http.StatusOK)
		assert.Equal(t, user.ID, body.User.ID)
	})

	t.Run("the email is case-insensitive", func(t *testing.T) {
		email := strings.ToUpper(user.Email)

		w := helper.Do(t, ts.Router, http.MethodPost, loginPath, `{"email":"`+email+`","password":"`+factory.Password+`"}`, "")

		assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
	})

	t.Run("wrong password is INVALID_CREDENTIALS", func(t *testing.T) {
		w := helper.Do(t, ts.Router, http.MethodPost, loginPath, dto.LoginRequest{Email: user.Email, Password: "wrong-password"}, "")

		assertError(t, w, http.StatusUnauthorized, helper.ErrorFrom(userdomain.ErrInvalidCredentials))
		assert.Nil(t, helper.Cookie(w, "refresh_token"))
	})

	t.Run("unknown email is INVALID_CREDENTIALS too", func(t *testing.T) {
		w := helper.Do(t, ts.Router, http.MethodPost, loginPath, dto.LoginRequest{Email: "nobody@example.com", Password: factory.Password}, "")

		assertError(t, w, http.StatusUnauthorized, helper.ErrorFrom(userdomain.ErrInvalidCredentials))
	})

	t.Run("missing fields are VALIDATION_FAILED", func(t *testing.T) {
		w := helper.Do(t, ts.Router, http.MethodPost, loginPath, `{}`, "")

		assertError(t, w, http.StatusBadRequest, helper.ErrorFrom(apperrors.Validation([]apperrors.Issue{
			{Field: "email", Message: "is required"},
			{Field: "password", Message: "is required"},
		})))
	})

	t.Run("factory.Login returns a working token and the cookie", func(t *testing.T) {
		accessToken, cookie := factory.Login(t, ts.Router, user.Email, factory.Password)

		assert.NotEmpty(t, cookie.Value)
		w := helper.Do(t, ts.Router, http.MethodGet, "/api/v1/users/me", nil, accessToken)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}
