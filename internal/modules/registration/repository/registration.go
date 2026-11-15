package repository

import (
	"context"
	"errors"
	"main/internal/infrastructure/database"
	"main/internal/modules/registration/domain"
	"main/internal/modules/registration/repository/mapper"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type registrationRepository struct {
	db *gorm.DB
}

func NewRegistrationRepository(db *gorm.DB) domain.Repository {
	return &registrationRepository{db: db}
}

func (r *registrationRepository) CreateWithQuotaDecrement(ctx context.Context, eventID, userID string) (*domain.Registration, error) {
	var created database.Registration

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var event database.Event
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", eventID).
			First(&event).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrEventNotFound
			}
			return err
		}

		if event.Status == "FINISHED" {
			return domain.ErrEventNotOpen
		}

		var existing database.Registration
		err := tx.Where("event_id = ? AND user_id = ?", eventID, userID).First(&existing).Error
		if err == nil {
			if existing.Status != string(domain.StatusCancelled) {
				return domain.ErrAlreadyRegistered
			}
			if event.AvailableQuota <= 0 {
				return domain.ErrEventQuotaFull
			}
			existing.Status = string(domain.StatusRegistered)
			if err := tx.Save(&existing).Error; err != nil {
				return err
			}
			if err := tx.Model(&event).
				Update("available_quota", gorm.Expr("available_quota - 1")).Error; err != nil {
				return err
			}
			created = existing
			return nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if event.AvailableQuota <= 0 {
			return domain.ErrEventQuotaFull
		}

		created = database.Registration{
			ID:      uuid.NewString(),
			EventID: eventID,
			UserID:  userID,
			Status:  string(domain.StatusRegistered),
		}
		if err := tx.Create(&created).Error; err != nil {
			return err
		}

		if err := tx.Model(&event).
			Update("available_quota", gorm.Expr("available_quota - 1")).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return r.FindByID(ctx, created.ID)
}

func (r *registrationRepository) FindByID(ctx context.Context, id string) (*domain.Registration, error) {
	var reg database.Registration
	if err := r.db.WithContext(ctx).
		Preload("User").
		Where("id = ?", id).
		First(&reg).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrRegistrationNotFound
		}
		return nil, err
	}
	return mapper.ToDomain(&reg), nil
}

func (r *registrationRepository) List(ctx context.Context, filter domain.ListFilter) ([]domain.Registration, int64, error) {
	query := r.db.WithContext(ctx).Model(&database.Registration{})

	if filter.UserID != "" {
		query = query.Where("user_id = ?", filter.UserID)
	}
	if filter.EventID != "" {
		query = query.Where("event_id = ?", filter.EventID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	var records []database.Registration
	if err := query.
		Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(filter.Limit).
		Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return mapper.ToDomainList(records), total, nil
}

func (r *registrationRepository) CancelWithQuotaIncrement(ctx context.Context, id string) (*domain.Registration, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var reg database.Registration
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", id).
			First(&reg).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrRegistrationNotFound
			}
			return err
		}

		switch domain.Status(reg.Status) {
		case domain.StatusCancelled:
			return domain.ErrAlreadyCancelled
		case domain.StatusCheckedIn:
			return domain.ErrCannotCancelCheckedIn
		}

		if err := tx.Model(&reg).Update("status", string(domain.StatusCancelled)).Error; err != nil {
			return err
		}

		if err := tx.Model(&database.Event{}).
			Where("id = ? AND available_quota < max_quota", reg.EventID).
			Update("available_quota", gorm.Expr("available_quota + 1")).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return r.FindByID(ctx, id)
}

func (r *registrationRepository) UpdateStatusWithQuotaSync(ctx context.Context, id string, newStatus domain.Status) (*domain.Registration, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var reg database.Registration
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", id).
			First(&reg).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrRegistrationNotFound
			}
			return err
		}

		oldStatus := domain.Status(reg.Status)
		if oldStatus == newStatus {
			return nil
		}

		var event database.Event
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", reg.EventID).
			First(&event).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrEventNotFound
			}
			return err
		}

		// Transitioning from active -> CANCELLED frees 1 slot.
		if oldStatus != domain.StatusCancelled && newStatus == domain.StatusCancelled {
			if event.AvailableQuota < event.MaxQuota {
				if err := tx.Model(&event).
					Update("available_quota", gorm.Expr("available_quota + 1")).Error; err != nil {
					return err
				}
			}
		}

		// Transitioning from CANCELLED -> active consumes 1 slot.
		if oldStatus == domain.StatusCancelled && newStatus != domain.StatusCancelled {
			if event.AvailableQuota <= 0 {
				return domain.ErrEventQuotaFull
			}
			if err := tx.Model(&event).
				Update("available_quota", gorm.Expr("available_quota - 1")).Error; err != nil {
				return err
			}
		}

		return tx.Model(&reg).Update("status", string(newStatus)).Error
	})
	if err != nil {
		return nil, err
	}

	return r.FindByID(ctx, id)
}

func (r *registrationRepository) DeleteWithQuotaSync(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var reg database.Registration
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", id).
			First(&reg).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return domain.ErrRegistrationNotFound
			}
			return err
		}

		if domain.Status(reg.Status) != domain.StatusCancelled {
			if err := tx.Model(&database.Event{}).
				Where("id = ? AND available_quota < max_quota", reg.EventID).
				Update("available_quota", gorm.Expr("available_quota + 1")).Error; err != nil {
				return err
			}
		}

		return tx.Delete(&reg).Error
	})
}

// CommitteeAuthorizer adapter backed by database.Committee (or replace with internal/shared/authz).
type committeeAuthorizer struct {
	db *gorm.DB
}

func NewCommitteeAuthorizer(db *gorm.DB) domain.CommitteeAuthorizer {
	return &committeeAuthorizer{db: db}
}

func (a *committeeAuthorizer) IsCoordinator(ctx context.Context, eventID, userID string) (bool, error) {
	var count int64
	err := a.db.WithContext(ctx).
		Model(&database.Committee{}).
		Where("event_id = ? AND user_id = ? AND role = ?", eventID, userID, "COORDINATOR").
		Count(&count).Error
	return count > 0, err
}

func (a *committeeAuthorizer) IsCommitteeMember(ctx context.Context, eventID, userID string) (bool, error) {
	var count int64
	err := a.db.WithContext(ctx).
		Model(&database.Committee{}).
		Where("event_id = ? AND user_id = ? AND role IN ?", eventID, userID, []string{"COORDINATOR", "STAFF"}).
		Count(&count).Error
	return count > 0, err
}
