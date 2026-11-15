package authz_test

import (
	"testing"

	"main/internal/shared/authz"
	apperrors "main/internal/shared/errors"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var errNotYours = apperrors.Forbidden("POST_NOT_OWNED", "not your post")

func TestHasRole(t *testing.T) {
	actor := authz.Actor{Role: authz.RoleUser}

	assert.True(t, actor.HasRole(authz.RoleAdmin, authz.RoleUser))
	assert.False(t, actor.HasRole(authz.RoleAdmin))
	assert.False(t, actor.HasRole())
}

func TestEnsureOwnerOrRole(t *testing.T) {
	owner := uuid.New()

	tests := []struct {
		name    string
		actor   authz.Actor
		ownerID uuid.UUID
		want    error
	}{
		{"owner allowed", authz.Actor{UserID: owner, Role: authz.RoleUser}, owner, nil},
		{"admin allowed", authz.Actor{UserID: uuid.New(), Role: authz.RoleAdmin}, owner, nil},
		{"other user gets the given error", authz.Actor{UserID: uuid.New(), Role: authz.RoleUser}, owner, errNotYours},
		{"empty actor never owns an empty id", authz.Actor{}, uuid.Nil, errNotYours},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authz.EnsureOwnerOrRole(tt.actor, tt.ownerID, errNotYours, authz.RoleAdmin)

			if tt.want == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.want)
			}
		})
	}
}
