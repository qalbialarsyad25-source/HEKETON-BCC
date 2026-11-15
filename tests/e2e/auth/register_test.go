package auth_test

import (
	"net/http"
	"strings"
	"testing"

	"main/internal/modules/auth/dto"
	userdomain "main/internal/modules/user/domain"
	apperrors "main/internal/shared/errors"
	"main/tests/e2e/helper"
	"main/tests/e2e/suite"

	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	ts := suite.New(t)

	t.Run("creates a user and signs in", func(t *testing.T) {
		req := dto.RegisterRequest{Name: "Alice", Email: "Alice@Example.COM", Password: "password123"}

		w := helper.Do(t, ts.Router, http.MethodPost, registerPath, req, "")

		body, _ := assertSessionResponse(t, ts, w, http.StatusCreated)
		assert.Equal(t, "Alice", body.User.Name)
		assert.Equal(t, "alice@example.com", body.User.Email)
		assert.NotContains(t, w.Body.String(), "password")
	})

	t.Run("the same email again is EMAIL_TAKEN", func(t *testing.T) {
		req := dto.RegisterRequest{Name: "Alice 2", Email: "ALICE@example.com", Password: "password123"}

		w := helper.Do(t, ts.Router, http.MethodPost, registerPath, req, "")

		assertError(t, w, http.StatusConflict, helper.ErrorFrom(userdomain.ErrEmailTaken))
	})

	t.Run("invalid body lists every issue", func(t *testing.T) {
		w := helper.Do(t, ts.Router, http.MethodPost, registerPath, `{"email":"not-an-email","password":"short"}`, "")

		assertError(t, w, http.StatusBadRequest, helper.ErrorFrom(apperrors.Validation([]apperrors.Issue{
			{Field: "name", Message: "is required"},
			{Field: "email", Message: "must be a valid email"},
			{Field: "password", Message: "must be at least 8 characters"},
		})))
	})

	t.Run("a password over 72 bytes is a validation error, not a 500", func(t *testing.T) {
		req := dto.RegisterRequest{Name: "Emoji", Email: "emoji@example.com", Password: strings.Repeat("é", 37)}

		w := helper.Do(t, ts.Router, http.MethodPost, registerPath, req, "")

		assertError(t, w, http.StatusBadRequest, helper.ErrorFrom(userdomain.ErrPasswordTooLong))
	})

	t.Run("malformed JSON is MALFORMED_BODY", func(t *testing.T) {
		w := helper.Do(t, ts.Router, http.MethodPost, registerPath, `{"name":`, "")

		assertError(t, w, http.StatusBadRequest, helper.ErrorFrom(apperrors.ErrMalformedBody))
	})
}
