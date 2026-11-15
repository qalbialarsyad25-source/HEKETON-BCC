// Package authz holds the ownership and role checks usecases run on the caller.
// It has no HTTP types, so usecases can be called from controllers, WebSocket handlers or jobs.
package authz

import (
	"slices"

	apperrors "main/internal/shared/errors"

	"github.com/google/uuid"
)

// Roles stored in users.role and in the JWT.
const (
	RoleAdmin = "admin"
	RoleUser  = "user"
)

// Actor is who performs an operation. Controllers get it with middleware.GetActor.
// This package must not import ports: ports uses Actor (realtime room authorizers).
type Actor struct {
	UserID uuid.UUID
	Role   string
}

// HasRole reports whether the actor has one of roles.
func (a Actor) HasRole(roles ...string) bool {
	return slices.Contains(roles, a.Role)
}

// EnsureOwnerOrRole returns nil when the actor owns the resource or has one of roles,
// otherwise notAllowed (usually a module error such as ErrPostNotOwned).
func EnsureOwnerOrRole(actor Actor, ownerID uuid.UUID, notAllowed *apperrors.AppError, roles ...string) error {
	if actor.UserID != uuid.Nil && actor.UserID == ownerID {
		return nil
	}
	if actor.HasRole(roles...) {
		return nil
	}
	return notAllowed
}
