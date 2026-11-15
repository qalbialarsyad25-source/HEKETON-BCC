package usecase

import (
	"context"
	"time"

	"main/internal/modules/event/domain"
	"main/internal/shared/pagination"

	"github.com/google/uuid"
)

// EventRepositoryI stores events. Event owns the events table and creates the coordinator's
// committee row together with a new event; everything else about committees belongs to the
// committee module.
type EventRepositoryI interface {
	// Create stores the event with coordinatorID as its coordinator and fills ID and timestamps.
	Create(ctx context.Context, event *domain.Event, coordinatorID uuid.UUID) error
	// FindByID returns nil, nil when no event has that id.
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Event, error)
	List(ctx context.Context, params pagination.Params) ([]domain.Event, int64, error)
	ListByMember(ctx context.Context, userID uuid.UUID, params pagination.Params) ([]domain.Event, int64, error)
	// Update overwrites the editable fields and moves the available quota with the max quota.
	// It returns nil, nil when the event is gone or the new max quota is below the number of
	// registered participants.
	Update(ctx context.Context, event *domain.Event) (*domain.Event, error)
	// Delete reports whether an event was deleted. Its sessions, registrations, announcements,
	// invitations and committee go with it (ON DELETE CASCADE).
	Delete(ctx context.Context, id uuid.UUID) (bool, error)
	IsCoordinator(ctx context.Context, eventID, userID uuid.UUID) (bool, error)
}

type EventUseCase struct {
	repo EventRepositoryI
}

func NewEventUseCase(repo EventRepositoryI) *EventUseCase {
	return &EventUseCase{repo: repo}
}

type CreateEventInput struct {
	UserID      uuid.UUID
	Title       string
	Description string
	Date        time.Time
	Location    string
	MaxQuota    int
}

// Create makes a new event, open for registration, and makes the creator its coordinator.
func (uc EventUseCase) Create(ctx context.Context, input CreateEventInput) (*domain.Event, error) {
	if !input.Date.After(time.Now()) {
		return nil, domain.ErrEventDateInPast
	}

	event := &domain.Event{
		Title:          input.Title,
		Description:    input.Description,
		Date:           input.Date,
		Location:       input.Location,
		AvailableQuota: input.MaxQuota,
		MaxQuota:       input.MaxQuota,
		Status:         domain.StatusUpcoming,
	}

	if err := uc.repo.Create(ctx, event, input.UserID); err != nil {
		return nil, err
	}

	return event, nil
}

func (uc EventUseCase) GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error) {
	event, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if event == nil {
		return nil, domain.ErrEventNotFound
	}

	return event, nil
}

func (uc EventUseCase) List(ctx context.Context, params pagination.Params) (*pagination.Result[domain.Event], error) {
	events, total, err := uc.repo.List(ctx, params)
	if err != nil {
		return nil, err
	}

	return pagination.NewResult(events, params, total), nil
}

// ListMine returns the events the user is a committee member of.
func (uc EventUseCase) ListMine(ctx context.Context, userID uuid.UUID, params pagination.Params) (*pagination.Result[domain.Event], error) {
	events, total, err := uc.repo.ListByMember(ctx, userID, params)
	if err != nil {
		return nil, err
	}

	return pagination.NewResult(events, params, total), nil
}

type UpdateEventInput struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Title       string
	Description string
	Date        time.Time
	Location    string
	MaxQuota    int
	Status      domain.Status // empty = keep the current status
}

// Update changes an event. Only a coordinator of the event may do it, and the max quota can't
// go below the number of participants who are already registered.
func (uc EventUseCase) Update(ctx context.Context, input UpdateEventInput) (*domain.Event, error) {
	event, err := uc.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if err := uc.requireCoordinator(ctx, input.ID, input.UserID); err != nil {
		return nil, err
	}

	event.Title = input.Title
	event.Description = input.Description
	event.Date = input.Date
	event.Location = input.Location
	event.MaxQuota = input.MaxQuota
	if input.Status != "" {
		event.Status = input.Status
	}

	updated, err := uc.repo.Update(ctx, event)
	if err != nil {
		return nil, err
	}

	if updated != nil {
		return updated, nil
	}

	current, err := uc.repo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if current == nil {
		return nil, domain.ErrEventNotFound
	}

	return nil, domain.ErrQuotaBelowRegistrations
}

// Delete removes an event with everything that belongs to it. Only a coordinator may do it.
func (uc EventUseCase) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if _, err := uc.GetByID(ctx, id); err != nil {
		return err
	}

	if err := uc.requireCoordinator(ctx, id, userID); err != nil {
		return err
	}

	deleted, err := uc.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	if !deleted {
		return domain.ErrEventNotFound
	}

	return nil
}

func (uc EventUseCase) requireCoordinator(ctx context.Context, eventID, userID uuid.UUID) error {
	ok, err := uc.repo.IsCoordinator(ctx, eventID, userID)
	if err != nil {
		return err
	}

	if !ok {
		return domain.ErrNotEventCoordinator
	}

	return nil
}
