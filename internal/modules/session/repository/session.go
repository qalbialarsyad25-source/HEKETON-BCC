package repository

import (
	"context"
	"errors"
	"fmt"

	"main/internal/infrastructure/database"
	"main/internal/modules/session/domain"
	"main/internal/modules/session/usecase"
	"main/internal/shared/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) usecase.SessionRepositoryI {
	return &SessionRepository{db: db}
}

// Create stores the session. On success session is filled with the stored timestamps.
func (repo SessionRepository) Create(ctx context.Context, session *domain.Session) error {
	row := toDatabaseSession(session)

	if err := repo.db.WithContext(ctx).Omit("Event").Create(row).Error; err != nil {
		return fmt.Errorf("session repository: create session: %w", err)
	}

	*session = *toDomainSession(row)

	return nil
}

func (repo SessionRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	row, err := gorm.G[database.Session](repo.db).
		Where("id = ?", id).
		First(ctx)

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("session repository: find session: %w", err)
	}

	return toDomainSession(&row), nil
}

// List counts the matching sessions and returns the requested page, earliest first.
// When eventID is set, only the sessions of that event.
func (repo SessionRepository) List(ctx context.Context, eventID *uuid.UUID, params pagination.Params) ([]domain.Session, int64, error) {
	query := repo.db.WithContext(ctx).Model(&database.Session{})
	if eventID != nil {
		query = query.Where("event_id = ?", *eventID)
	}
	query = query.Session(&gorm.Session{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("session repository: count sessions: %w", err)
	}

	var rows []database.Session
	err := query.
		Order(`"time" ASC, id ASC`).
		Limit(params.Limit).
		Offset(params.Offset()).
		Find(&rows).Error
	if err != nil {
		return nil, 0, fmt.Errorf("session repository: list sessions: %w", err)
	}

	sessions := make([]domain.Session, 0, len(rows))
	for i := range rows {
		sessions = append(sessions, *toDomainSession(&rows[i]))
	}

	return sessions, total, nil
}

// Update overwrites the editable fields in one UPDATE. It returns nil, nil when no row matched:
// the session is gone.
func (repo SessionRepository) Update(ctx context.Context, session *domain.Session) (*domain.Session, error) {
	result := repo.db.WithContext(ctx).
		Model(&database.Session{}).
		Where("id = ?", session.ID).
		Updates(map[string]any{
			"title":    session.Title,
			"time":     session.Time,
			"speaker":  session.Speaker,
			"location": session.Location,
			"status":   string(session.Status),
		})
	if result.Error != nil {
		return nil, fmt.Errorf("session repository: update session: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, nil
	}

	return repo.FindByID(ctx, session.ID)
}

func (repo SessionRepository) Delete(ctx context.Context, id uuid.UUID) (bool, error) {
	deletedRows, err := gorm.G[database.Session](repo.db).
		Where("id = ?", id).
		Delete(ctx)

	if err != nil {
		return false, fmt.Errorf("session repository: delete session: %w", err)
	}

	return deletedRows == 1, nil
}

func (repo SessionRepository) EventExists(ctx context.Context, eventID uuid.UUID) (bool, error) {
	return repo.exists(ctx, &database.Event{}, "event exists", "id = ?", eventID)
}

func (repo SessionRepository) IsCommitteeMember(ctx context.Context, eventID, userID uuid.UUID) (bool, error) {
	return repo.exists(ctx, &database.Committee{}, "check committee member",
		"event_id = ? AND user_id = ?", eventID, userID)
}

func (repo SessionRepository) exists(ctx context.Context, model any, what, query string, args ...any) (bool, error) {
	var count int64

	err := repo.db.WithContext(ctx).
		Model(model).
		Where(query, args...).
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("session repository: %s: %w", what, err)
	}

	return count > 0, nil
}

func toDatabaseSession(s *domain.Session) *database.Session {
	return &database.Session{
		ID:        s.ID,
		EventID:   s.EventID,
		Title:     s.Title,
		Time:      s.Time,
		Speaker:   s.Speaker,
		Location:  s.Location,
		Status:    string(s.Status),
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

func toDomainSession(row *database.Session) *domain.Session {
	return &domain.Session{
		ID:        row.ID,
		EventID:   row.EventID,
		Title:     row.Title,
		Time:      row.Time,
		Speaker:   row.Speaker,
		Location:  row.Location,
		Status:    domain.Status(row.Status),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
