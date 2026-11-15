package repository

import (
	"context"
	"errors"
	"main/internal/modules/announcement/domain"
	"main/internal/modules/announcement/usecase"
	"main/internal/infrastructure/database"
	"main/internal/modules/announcement/repository/mapper"
	"main/internal/shared/pagination"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AnnouncementRepository struct {
	db *gorm.DB
}

func NewAnnouncementRepository(db *gorm.DB) usecase.AnnouncementRepositoryI {
	return &AnnouncementRepository{db: db}
}

func (u *AnnouncementRepository) GetAnnouncement(ctx context.Context, page pagination.Pagination) ([]domain.Announcement, error) {
	announcements, err := gorm.G[database.Announcement](u.db).
		Order("created_at DESC").
		Limit(page.Limit).
		Offset(page.Offset).
		Find(ctx)
	if err != nil {
		return nil, err
	}

	return mapper.ToDomainAnnouncementBatch(announcements), nil
}

func (u *AnnouncementRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Announcement, error) {
	announcement, err := gorm.G[database.Announcement](u.db).
		Where("id = ?", id).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrAnnouncementNotFound
		}
		return nil, err
	}

	return mapper.ToDomainAnnouncement(&announcement), nil
}

func (u *AnnouncementRepository) UpdateAnnouncement(ctx context.Context, announcement *domain.Announcement) error {
	res := u.db.WithContext(ctx).
		Model(&database.Announcement{}).
		Where("id = ?", announcement.ID).
		Updates(map[string]any{
			"title": announcement.Title,
			"body":  announcement.Body,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domain.ErrAnnouncementNotFound
	}

	return nil
}

func (u *AnnouncementRepository) DeleteAnnouncement(ctx context.Context, id uuid.UUID) error {
	rows, err := gorm.G[database.Announcement](u.db).
		Where("id = ?", id).
		Delete(ctx)
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrAnnouncementNotFound
	}

	return nil
}
