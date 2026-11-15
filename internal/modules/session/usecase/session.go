package usecase

import (
	"context"
	"time"

	"main/internal/modules/session/domain"
	"main/internal/shared/pagination"

	"github.com/google/uuid"
)

// SessionRepositoryI is what the usecase needs from storage. Find-style methods and the
// conditional update return nil, nil when no row matched.
type SessionRepositoryI interface {
	Create(ctx context.Context, session *domain.Session) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Session, error)
	// List returns the sessions of one event when eventID is set, else of every event.
	List(ctx context.Context, eventID *uuid.UUID, params pagination.Params) ([]domain.Session, int64, error)
	// Update overwrites the editable fields. It returns nil, nil when the session is gone.
	Update(ctx context.Context, session *domain.Session) (*domain.Session, error)
	Delete(ctx context.Context, id uuid.UUID) (bool, error)

	EventExists(ctx context.Context, eventID uuid.UUID) (bool, error)
	// IsCommitteeMember reports whether the user is on the event's committee, as COORDINATOR or STAFF.
	IsCommitteeMember(ctx context.Context, eventID, userID uuid.UUID) (bool, error)
}

// Notifier tells clients that follow an event about changes to its sessions.
// It is optional (nil means nobody is told) and must never block or fail the request.
type Notifier interface {
	SessionCreated(session domain.Session)
	SessionUpdated(session domain.Session)
	SessionDeleted(session domain.Session)
}

type SessionUseCase struct {
	repo     SessionRepositoryI
	notifier Notifier
}

func NewSessionUseCase(repo SessionRepositoryI, notifier Notifier) *SessionUseCase {
	return &SessionUseCase{repo: repo, notifier: notifier}
}

type CreateSessionInput struct {
	UserID   uuid.UUID
	EventID  uuid.UUID
	Title    string
	Time     time.Time
	Speaker  string
	Location string
}

// Create adds an UPCOMING session to an event. Only a committee member of the event can do it.
func (uc SessionUseCase) Create(ctx context.Context, input CreateSessionInput) (*domain.Session, error) {
	eventExists, err := uc.repo.EventExists(ctx, input.EventID)
	if err != nil {
		return nil, err
	}

	if !eventExists {
		return nil, domain.ErrEventNotFound
	}

	if err := uc.requireCommitteeMember(ctx, input.EventID, input.UserID); err != nil {
		return nil, err
	}

	session := &domain.Session{
		ID:       uuid.New(),
		EventID:  input.EventID,
		Title:    input.Title,
		Time:     input.Time,
		Speaker:  input.Speaker,
		Location: input.Location,
		Status:   domain.StatusUpcoming,
	}

	if err := uc.repo.Create(ctx, session); err != nil {
		return nil, err
	}

	if uc.notifier != nil {
		uc.notifier.SessionCreated(*session)
	}

	return session, nil
}

func (uc SessionUseCase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	return uc.find(ctx, id)
}

// List returns the sessions of one event when eventID is set, else of every event, earliest first.
func (uc SessionUseCase) List(ctx context.Context, eventID *uuid.UUID, params pagination.Params) (*pagination.Result[domain.Session], error) {
	sessions, total, err := uc.repo.List(ctx, eventID, params)
	if err != nil {
		return nil, err
	}

	return pagination.NewResult(sessions, params, total), nil
}

type UpdateSessionInput struct {
	ID       uuid.UUID
	UserID   uuid.UUID
	Title    string
	Time     time.Time
	Speaker  string
	Location string
	Status   domain.Status // empty = keep the current status
}

// Update changes a session. Only a committee member of its event can do it.
func (uc SessionUseCase) Update(ctx context.Context, input UpdateSessionInput) (*domain.Session, error) {
	if input.Status != "" && !input.Status.Valid() {
		return nil, domain.ErrInvalidStatus
	}

	session, err := uc.find(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if err := uc.requireCommitteeMember(ctx, session.EventID, input.UserID); err != nil {
		return nil, err
	}

	session.Title = input.Title
	session.Time = input.Time
	session.Speaker = input.Speaker
	session.Location = input.Location
	if input.Status != "" {
		session.Status = input.Status
	}

	updated, err := uc.repo.Update(ctx, session)
	if err != nil {
		return nil, err
	}

	if updated == nil {
		return nil, domain.ErrSessionNotFound
	}

	if uc.notifier != nil {
		uc.notifier.SessionUpdated(*updated)
	}

	return updated, nil
}

// Delete removes a session and returns it as it was. Only a committee member of its event can do it.
func (uc SessionUseCase) Delete(ctx context.Context, id, userID uuid.UUID) (*domain.Session, error) {
	session, err := uc.find(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := uc.requireCommitteeMember(ctx, session.EventID, userID); err != nil {
		return nil, err
	}

	deleted, err := uc.repo.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	if !deleted {
		return nil, domain.ErrSessionNotFound
	}

	if uc.notifier != nil {
		uc.notifier.SessionDeleted(*session)
	}

	return session, nil
}

// EventExists lets the controller refuse a realtime subscription to an event that isn't there.
func (uc SessionUseCase) EventExists(ctx context.Context, eventID uuid.UUID) (bool, error) {
	return uc.repo.EventExists(ctx, eventID)
}

func (uc SessionUseCase) find(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	session, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if session == nil {
		return nil, domain.ErrSessionNotFound
	}

	return session, nil
}

func (uc SessionUseCase) requireCommitteeMember(ctx context.Context, eventID, userID uuid.UUID) error {
	member, err := uc.repo.IsCommitteeMember(ctx, eventID, userID)
	if err != nil {
		return err
	}

	if !member {
		return domain.ErrNotEventMember
	}

	return nil
}
