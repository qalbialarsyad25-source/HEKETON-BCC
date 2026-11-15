package repository

import (
	"context"
	"errors"
	"fmt"

	"main/internal/infrastructure/database"
	"main/internal/modules/event/domain"
	"main/internal/modules/event/repository/mapper"
	"main/internal/modules/event/usecase"
	"main/internal/shared/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EventRepository struct {
	db *gorm.DB
}

func NewEventRepository(db *gorm.DB) usecase.EventRepositoryI {
	return &EventRepository{db: db}
}

// Create stores the event and makes coordinatorID its coordinator in one transaction.
// On success event is filled with the stored ID and timestamps.
func (repo EventRepository) Create(ctx context.Context, event *domain.Event, coordinatorID uuid.UUID) error {
	row := mapper.ToDatabaseEvent(event)

	err := repo.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := gorm.G[database.Event](tx).Create(ctx, row); err != nil {
			return err
		}

		member := database.Committee{
			EventID: row.ID,
			UserID:  coordinatorID,
			Role:    database.RoleCoordinator,
		}

		return gorm.G[database.Committee](tx).Create(ctx, &member)
	})
	if err != nil {
		return fmt.Errorf("event repository: create event: %w", err)
	}

	*event = *mapper.ToDomainEvent(row)

	return nil
}

func (repo EventRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Event, error) {
	row, err := gorm.G[database.Event](repo.db).
		Where("id = ?", id).
		First(ctx)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("event repository: find event: %w", err)
	}

	return mapper.ToDomainEvent(&row), nil
}

func (repo EventRepository) List(ctx context.Context, params pagination.Params) ([]domain.Event, int64, error) {
	events, total, err := repo.page(ctx, params, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("event repository: list events: %w", err)
	}

	return events, total, nil
}

// ListByMember returns the events the user is a committee member of.
func (repo EventRepository) ListByMember(ctx context.Context, userID uuid.UUID, params pagination.Params) ([]domain.Event, int64, error) {
	memberOf := func(query *gorm.DB) *gorm.DB {
		events := repo.db.Model(&database.Committee{}).
			Select("event_id").
			Where("user_id = ?", userID)

		return query.Where("id IN (?)", events)
	}

	events, total, err := repo.page(ctx, params, memberOf)
	if err != nil {
		return nil, 0, fmt.Errorf("event repository: list events of member: %w", err)
	}

	return events, total, nil
}

// Update overwrites the editable fields in one UPDATE. The available quota moves with the max
// quota, and the WHERE refuses a max quota below the number of registered participants, so
// a registration that lands during the update can't push the quota negative. It returns nil, nil
// when no row matched: the event is gone, or the new max quota is too low.
func (repo EventRepository) Update(ctx context.Context, event *domain.Event) (*domain.Event, error) {
	result := repo.db.WithContext(ctx).
		Model(&database.Event{}).
		Where("id = ? AND (max_quota - available_quota) <= ?", event.ID, event.MaxQuota).
		Updates(map[string]any{
			"title":           event.Title,
			"description":     event.Description,
			"date":            event.Date,
			"location":        event.Location,
			"status":          string(event.Status),
			"max_quota":       event.MaxQuota,
			"available_quota": gorm.Expr("? - (max_quota - available_quota)", event.MaxQuota),
		})
	if result.Error != nil {
		return nil, fmt.Errorf("event repository: update event: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return repo.FindByID(ctx, event.ID)
}

func (repo EventRepository) Delete(ctx context.Context, id uuid.UUID) (bool, error) {
	deletedRows, err := gorm.G[database.Event](repo.db).
		Where("id = ?", id).
		Delete(ctx)

	if err != nil {
		return false, fmt.Errorf("event repository: delete event: %w", err)
	}

	return deletedRows == 1, nil
}

func (repo EventRepository) IsCoordinator(ctx context.Context, eventID, userID uuid.UUID) (bool, error) {
	count, err := gorm.G[database.Committee](repo.db).
		Where("event_id = ? AND user_id = ? AND role = ?", eventID, userID, database.RoleCoordinator).
		Count(ctx, "*")

	if err != nil {
		return false, fmt.Errorf("event repository: check coordinator: %w", err)
	}

	return count > 0, nil
}

// page counts the events that match scope and returns the requested page, latest date first.
func (repo EventRepository) page(ctx context.Context, params pagination.Params, scope func(*gorm.DB) *gorm.DB) ([]domain.Event, int64, error) {
	query := repo.db.WithContext(ctx).Model(&database.Event{})
	if scope != nil {
		query = scope(query)
	}
	query = query.Session(&gorm.Session{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []database.Event
	err := query.
		Order("date DESC, id").
		Limit(params.Limit).
		Offset(params.Offset()).
		Find(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	return mapper.ToDomainEvents(rows), total, nil
}
