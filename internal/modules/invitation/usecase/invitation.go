package usecase

import (
	"context"

	"github.com/google/uuid"

	"main/internal/modules/invitation/domain"
	"main/internal/shared/pagination"
)

// InvitationRepositoryI is what the usecase needs from storage. Find-style methods and the
// conditional updates return nil, nil when no row matched.
type InvitationRepositoryI interface {
	Create(ctx context.Context, invitation *domain.Invitation) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Invitation, error)
	ListByCoordinator(ctx context.Context, userID uuid.UUID, eventID *uuid.UUID, params pagination.Params) ([]domain.Invitation, int64, error)
	ListByReceiver(ctx context.Context, userID uuid.UUID, params pagination.Params) ([]domain.Invitation, int64, error)
	// UpdateRole changes the role of a PENDING invitation.
	UpdateRole(ctx context.Context, id uuid.UUID, role domain.Role) (*domain.Invitation, error)
	// Transition moves an invitation from one status to another, only if it is still in from.
	Transition(ctx context.Context, id uuid.UUID, from, to domain.Status) (*domain.Invitation, error)
	// Accept marks a PENDING invitation ACCEPTED and adds the receiver to the event's committee, in one transaction.
	Accept(ctx context.Context, id uuid.UUID) (*domain.Invitation, error)
	Delete(ctx context.Context, id uuid.UUID) (bool, error)

	EventExists(ctx context.Context, eventID uuid.UUID) (bool, error)
	UserExists(ctx context.Context, userID uuid.UUID) (bool, error)
	IsCoordinator(ctx context.Context, eventID, userID uuid.UUID) (bool, error)
	IsMember(ctx context.Context, eventID, userID uuid.UUID) (bool, error)
	HasPending(ctx context.Context, eventID, receiverID uuid.UUID) (bool, error)
}

type CreateInvitationInput struct {
	SenderID   uuid.UUID
	ReceiverID uuid.UUID
	EventID    uuid.UUID
	Role       domain.Role
}

type UpdateInvitationInput struct {
	ID     uuid.UUID
	UserID uuid.UUID
	Role   domain.Role
}

type InvitationUseCase struct {
	repo InvitationRepositoryI
}

func NewInvitationUseCase(repo InvitationRepositoryI) *InvitationUseCase {
	return &InvitationUseCase{repo: repo}
}

// Create sends an invitation. Only a coordinator of the event can invite, and only a user
// who isn't on the committee yet and has no pending invitation to the event.
func (uc InvitationUseCase) Create(ctx context.Context, input CreateInvitationInput) (*domain.Invitation, error) {
	if !input.Role.Valid() {
		return nil, domain.ErrInvalidRole
	}

	if input.ReceiverID == input.SenderID {
		return nil, domain.ErrCannotInviteSelf
	}

	eventExists, err := uc.repo.EventExists(ctx, input.EventID)
	if err != nil {
		return nil, err
	}

	if !eventExists {
		return nil, domain.ErrEventNotFound
	}

	if err := uc.requireCoordinator(ctx, input.EventID, input.SenderID); err != nil {
		return nil, err
	}

	receiverExists, err := uc.repo.UserExists(ctx, input.ReceiverID)
	if err != nil {
		return nil, err
	}

	if !receiverExists {
		return nil, domain.ErrReceiverNotFound
	}

	member, err := uc.repo.IsMember(ctx, input.EventID, input.ReceiverID)
	if err != nil {
		return nil, err
	}

	if member {
		return nil, domain.ErrAlreadyMember
	}

	pending, err := uc.repo.HasPending(ctx, input.EventID, input.ReceiverID)
	if err != nil {
		return nil, err
	}

	if pending {
		return nil, domain.ErrInvitationAlreadyPending
	}

	invitation := &domain.Invitation{
		EventID:    input.EventID,
		Sender:     domain.Sender{ID: input.SenderID},
		ReceiverID: input.ReceiverID,
		Role:       input.Role,
		Status:     domain.StatusPending,
	}

	if err := uc.repo.Create(ctx, invitation); err != nil {
		return nil, err
	}

	return invitation, nil
}

// GetByID returns the invitation to its receiver, its sender and the coordinators of its event.
// Anyone else gets a not-found, so ids can't be probed.
func (uc InvitationUseCase) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Invitation, error) {
	invitation, err := uc.find(ctx, id)
	if err != nil {
		return nil, err
	}

	if invitation.ReceiverID == userID || invitation.Sender.ID == userID {
		return invitation, nil
	}

	coordinator, err := uc.repo.IsCoordinator(ctx, invitation.EventID, userID)
	if err != nil {
		return nil, err
	}

	if !coordinator {
		return nil, domain.ErrInvitationNotFound
	}

	return invitation, nil
}

// List returns the invitations of the events the user coordinates, optionally of one event.
func (uc InvitationUseCase) List(ctx context.Context, userID uuid.UUID, eventID *uuid.UUID, params pagination.Params) (*pagination.Result[domain.Invitation], error) {
	items, total, err := uc.repo.ListByCoordinator(ctx, userID, eventID, params)
	if err != nil {
		return nil, err
	}

	return pagination.NewResult(items, params, total), nil
}

// ListMine returns the invitations sent to the user.
func (uc InvitationUseCase) ListMine(ctx context.Context, userID uuid.UUID, params pagination.Params) (*pagination.Result[domain.Invitation], error) {
	items, total, err := uc.repo.ListByReceiver(ctx, userID, params)
	if err != nil {
		return nil, err
	}

	return pagination.NewResult(items, params, total), nil
}

// UpdateRole changes the role offered by a pending invitation. Coordinators only.
func (uc InvitationUseCase) UpdateRole(ctx context.Context, input UpdateInvitationInput) (*domain.Invitation, error) {
	if !input.Role.Valid() {
		return nil, domain.ErrInvalidRole
	}

	invitation, err := uc.find(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if err := uc.requireCoordinator(ctx, invitation.EventID, input.UserID); err != nil {
		return nil, err
	}

	updated, err := uc.repo.UpdateRole(ctx, input.ID, input.Role)
	if err != nil {
		return nil, err
	}

	if updated == nil {
		return nil, domain.ErrInvitationNotPending
	}

	return updated, nil
}

// Delete removes the invitation record. Coordinators only. It doesn't remove anyone from the committee.
func (uc InvitationUseCase) Delete(ctx context.Context, id, userID uuid.UUID) error {
	invitation, err := uc.find(ctx, id)
	if err != nil {
		return err
	}

	if err := uc.requireCoordinator(ctx, invitation.EventID, userID); err != nil {
		return err
	}

	deleted, err := uc.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	if !deleted {
		return domain.ErrInvitationNotFound
	}

	return nil
}

// Accept lets the receiver join the committee with the invited role.
func (uc InvitationUseCase) Accept(ctx context.Context, id, userID uuid.UUID) (*domain.Invitation, error) {
	invitation, err := uc.findForReceiver(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	member, err := uc.repo.IsMember(ctx, invitation.EventID, userID)
	if err != nil {
		return nil, err
	}

	if member {
		return nil, domain.ErrAlreadyMember
	}

	accepted, err := uc.repo.Accept(ctx, id)
	if err != nil {
		return nil, err
	}

	if accepted == nil {
		return nil, domain.ErrInvitationNotPending
	}

	return accepted, nil
}

// Decline lets the receiver turn the invitation down.
func (uc InvitationUseCase) Decline(ctx context.Context, id, userID uuid.UUID) (*domain.Invitation, error) {
	if _, err := uc.findForReceiver(ctx, id, userID); err != nil {
		return nil, err
	}

	return uc.transition(ctx, id, domain.StatusDeclined)
}

// Cancel withdraws a pending invitation. Coordinators of the event only.
func (uc InvitationUseCase) Cancel(ctx context.Context, id, userID uuid.UUID) (*domain.Invitation, error) {
	invitation, err := uc.find(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := uc.requireCoordinator(ctx, invitation.EventID, userID); err != nil {
		return nil, err
	}

	return uc.transition(ctx, id, domain.StatusCancelled)
}

func (uc InvitationUseCase) transition(ctx context.Context, id uuid.UUID, to domain.Status) (*domain.Invitation, error) {
	updated, err := uc.repo.Transition(ctx, id, domain.StatusPending, to)
	if err != nil {
		return nil, err
	}

	if updated == nil {
		return nil, domain.ErrInvitationNotPending
	}

	return updated, nil
}

func (uc InvitationUseCase) find(ctx context.Context, id uuid.UUID) (*domain.Invitation, error) {
	invitation, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if invitation == nil {
		return nil, domain.ErrInvitationNotFound
	}

	return invitation, nil
}

// findForReceiver finds the invitation for its receiver. Anyone else gets a not-found.
func (uc InvitationUseCase) findForReceiver(ctx context.Context, id, userID uuid.UUID) (*domain.Invitation, error) {
	invitation, err := uc.find(ctx, id)
	if err != nil {
		return nil, err
	}

	if invitation.ReceiverID != userID {
		return nil, domain.ErrInvitationNotFound
	}

	return invitation, nil
}

func (uc InvitationUseCase) requireCoordinator(ctx context.Context, eventID, userID uuid.UUID) error {
	coordinator, err := uc.repo.IsCoordinator(ctx, eventID, userID)
	if err != nil {
		return err
	}

	if !coordinator {
		return domain.ErrNotEventCoordinator
	}

	return nil
}
