package repository

import (
	"context"
	"errors"
	"fmt"

	"main/internal/infrastructure/database"
	"main/internal/modules/invitation/domain"
	"main/internal/modules/invitation/usecase"
	"main/internal/shared/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var errNotPending = errors.New("invitation is not pending")

type InvitationRepository struct {
	db *gorm.DB
}

func NewInvitationRepository(db *gorm.DB) usecase.InvitationRepositoryI {
	return &InvitationRepository{db: db}
}

// Create stores the invitation. On success invitation is filled with the stored ID, timestamps and sender.
func (repo InvitationRepository) Create(ctx context.Context, invitation *domain.Invitation) error {
	row := toDatabaseInvitation(invitation)

	if err := repo.db.WithContext(ctx).Omit("Sender", "Event").Create(row).Error; err != nil {
		return fmt.Errorf("invitation repository: create invitation: %w", err)
	}

	stored, err := repo.FindByID(ctx, row.ID)
	if err != nil {
		return err
	}

	if stored == nil {
		return errors.New("invitation repository: create invitation: stored invitation not found")
	}

	*invitation = *stored

	return nil
}

func (repo InvitationRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Invitation, error) {
	var row database.Invitation

	err := repo.db.WithContext(ctx).
		Preload("Sender").
		Where("id = ?", id).
		First(&row).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("invitation repository: find invitation: %w", err)
	}

	return toDomainInvitation(&row), nil
}

// ListByCoordinator returns the invitations of the events the user coordinates, newest first.
// When eventID is set, only the invitations of that event.
func (repo InvitationRepository) ListByCoordinator(ctx context.Context, userID uuid.UUID, eventID *uuid.UUID, params pagination.Params) ([]domain.Invitation, int64, error) {
	coordinates := func(query *gorm.DB) *gorm.DB {
		events := repo.db.Model(&database.Committee{}).
			Select("event_id").
			Where("user_id = ? AND role = ?", userID, database.RoleCoordinator)

		query = query.Where("event_id IN (?)", events)
		if eventID != nil {
			query = query.Where("event_id = ?", *eventID)
		}

		return query
	}

	invitations, total, err := repo.page(ctx, params, coordinates)
	if err != nil {
		return nil, 0, fmt.Errorf("invitation repository: list invitations: %w", err)
	}

	return invitations, total, nil
}

// ListByReceiver returns the invitations sent to the user, newest first.
func (repo InvitationRepository) ListByReceiver(ctx context.Context, userID uuid.UUID, params pagination.Params) ([]domain.Invitation, int64, error) {
	received := func(query *gorm.DB) *gorm.DB {
		return query.Where("receiver_id = ?", userID)
	}

	invitations, total, err := repo.page(ctx, params, received)
	if err != nil {
		return nil, 0, fmt.Errorf("invitation repository: list invitations of receiver: %w", err)
	}

	return invitations, total, nil
}

// UpdateRole changes the role in one UPDATE that only matches a PENDING invitation.
// It returns nil, nil when no row matched: the invitation is gone or already answered.
func (repo InvitationRepository) UpdateRole(ctx context.Context, id uuid.UUID, role domain.Role) (*domain.Invitation, error) {
	result := repo.db.WithContext(ctx).
		Model(&database.Invitation{}).
		Where("id = ? AND status = ?", id, string(domain.StatusPending)).
		Update("role", string(role))
	if result.Error != nil {
		return nil, fmt.Errorf("invitation repository: update role: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return repo.FindByID(ctx, id)
}

// Transition moves the invitation from one status to another in one UPDATE, so two answers
// that arrive together can't both win. It returns nil, nil when the invitation isn't in from.
func (repo InvitationRepository) Transition(ctx context.Context, id uuid.UUID, from, to domain.Status) (*domain.Invitation, error) {
	result := repo.db.WithContext(ctx).
		Model(&database.Invitation{}).
		Where("id = ? AND status = ?", id, string(from)).
		Update("status", string(to))
	if result.Error != nil {
		return nil, fmt.Errorf("invitation repository: change status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return repo.FindByID(ctx, id)
}

// Accept marks the invitation ACCEPTED and adds the receiver to the event's committee in one
// transaction. It returns nil, nil when the invitation is no longer PENDING.
func (repo InvitationRepository) Accept(ctx context.Context, id uuid.UUID) (*domain.Invitation, error) {
	err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&database.Invitation{}).
			Where("id = ? AND status = ?", id, string(domain.StatusPending)).
			Update("status", string(domain.StatusAccepted))
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errNotPending
		}

		var row database.Invitation
		if err := tx.Where("id = ?", id).First(&row).Error; err != nil {
			return err
		}

		member := database.Committee{
			EventID: row.EventID,
			UserID:  row.ReceiverID,
		}

		if domain.Role(row.Role) == domain.RoleCoordinator {
			member.Role = database.RoleCoordinator
		} else {
			member.Role = database.RoleStaff
		}

		return tx.Create(&member).Error
	})

	if errors.Is(err, errNotPending) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("invitation repository: accept invitation: %w", err)
	}

	return repo.FindByID(ctx, id)
}

func (repo InvitationRepository) Delete(ctx context.Context, id uuid.UUID) (bool, error) {
	result := repo.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&database.Invitation{})
	if result.Error != nil {
		return false, fmt.Errorf("invitation repository: delete invitation: %w", result.Error)
	}

	return result.RowsAffected == 1, nil
}

func (repo InvitationRepository) EventExists(ctx context.Context, eventID uuid.UUID) (bool, error) {
	return repo.exists(ctx, &database.Event{}, "event exists", "id = ?", eventID)
}

func (repo InvitationRepository) UserExists(ctx context.Context, userID uuid.UUID) (bool, error) {
	return repo.exists(ctx, &database.User{}, "user exists", "id = ?", userID)
}

func (repo InvitationRepository) IsCoordinator(ctx context.Context, eventID, userID uuid.UUID) (bool, error) {
	return repo.exists(ctx, &database.Committee{}, "check coordinator",
		"event_id = ? AND user_id = ? AND role = ?", eventID, userID, database.RoleCoordinator)
}

func (repo InvitationRepository) IsMember(ctx context.Context, eventID, userID uuid.UUID) (bool, error) {
	return repo.exists(ctx, &database.Committee{}, "check member",
		"event_id = ? AND user_id = ?", eventID, userID)
}

func (repo InvitationRepository) HasPending(ctx context.Context, eventID, receiverID uuid.UUID) (bool, error) {
	return repo.exists(ctx, &database.Invitation{}, "check pending",
		"event_id = ? AND receiver_id = ? AND status = ?", eventID, receiverID, string(domain.StatusPending))
}

func (repo InvitationRepository) exists(ctx context.Context, model any, what, query string, args ...any) (bool, error) {
	var count int64

	err := repo.db.WithContext(ctx).
		Model(model).
		Where(query, args...).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("invitation repository: %s: %w", what, err)
	}

	return count > 0, nil
}

// page counts the invitations that match scope and returns the requested page, newest first.
func (repo InvitationRepository) page(ctx context.Context, params pagination.Params, scope func(*gorm.DB) *gorm.DB) ([]domain.Invitation, int64, error) {
	query := repo.db.WithContext(ctx).Model(&database.Invitation{})
	if scope != nil {
		query = scope(query)
	}
	query = query.Session(&gorm.Session{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []database.Invitation
	err := query.
		Preload("Sender").
		Order("created_at DESC, id").
		Limit(params.Limit).
		Offset(params.Offset()).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	invitations := make([]domain.Invitation, 0, len(rows))
	for i := range rows {
		invitations = append(invitations, *toDomainInvitation(&rows[i]))
	}

	return invitations, total, nil
}

func toDatabaseInvitation(invitation *domain.Invitation) *database.Invitation {
	return &database.Invitation{
		ID:         invitation.ID,
		EventID:    invitation.EventID,
		SenderID:   invitation.Sender.ID,
		ReceiverID: invitation.ReceiverID,
		Role:       string(invitation.Role),
		Status:     string(invitation.Status),
	}
}

func toDomainInvitation(row *database.Invitation) *domain.Invitation {
	return &domain.Invitation{
		ID:      row.ID,
		EventID: row.EventID,
		Sender: domain.Sender{
			ID:    row.SenderID,
			Email: row.Sender.Email,
		},
		ReceiverID: row.ReceiverID,
		Role:       domain.Role(row.Role),
		Status:     domain.Status(row.Status),
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
	}
}
