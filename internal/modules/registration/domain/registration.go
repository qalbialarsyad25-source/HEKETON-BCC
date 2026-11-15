package domain

import (
	"context"
	"time"
)

type Status string

const (
	StatusRegistered Status = "REGISTERED"
	StatusCheckedIn  Status = "CHECKEDIN"
	StatusCancelled  Status = "CANCELLED"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusRegistered, StatusCheckedIn, StatusCancelled:
		return true
	default:
		return false
	}
}

type UserSummary struct {
	ID    string
	Email string
	Name  string
}

type Registration struct {
	ID        string
	EventID   string
	UserID    string
	Status    Status
	User      UserSummary
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ListFilter struct {
	UserID  string
	EventID string
	Page    int
	Limit   int
}

type Repository interface {
	CreateWithQuotaDecrement(ctx context.Context, eventID, userID string) (*Registration, error)
	FindByID(ctx context.Context, id string) (*Registration, error)
	List(ctx context.Context, filter ListFilter) ([]Registration, int64, error)
	CancelWithQuotaIncrement(ctx context.Context, id string) (*Registration, error)
	UpdateStatusWithQuotaSync(ctx context.Context, id string, newStatus Status) (*Registration, error)
	DeleteWithQuotaSync(ctx context.Context, id string) error
}

type CommitteeAuthorizer interface {
	IsCoordinator(ctx context.Context, eventID, userID string) (bool, error)
	IsCommitteeMember(ctx context.Context, eventID, userID string) (bool, error)
}

type Usecase interface {
	RegisterEvent(ctx context.Context, eventID, userID string) (*Registration, error)
	GetMyRegistrations(ctx context.Context, userID string, page, limit int) ([]Registration, int64, error)
	GetRegistrations(ctx context.Context, eventID string, page, limit int) ([]Registration, int64, error)
	CancelRegistration(ctx context.Context, registrationID, actorUserID string) (*Registration, error)
	UpdateRegistration(ctx context.Context, registrationID, actorUserID string, status Status) (*Registration, error)
	DeleteRegistration(ctx context.Context, registrationID, actorUserID string) error
	CheckInRegistration(ctx context.Context, registrationID, actorUserID string) (*Registration, error)
}
