package usecase

import (
	"context"

	"main/internal/modules/announcement/domain"
	shared "main/internal/shared/pagination"

	"github.com/google/uuid"
)

type AnnouncementRepositoryI interface {
	GetAnnouncement(ctx context.Context, page shared.Pagination) ([]domain.Announcement, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Announcement, error)
	UpdateAnnouncement(ctx context.Context, announcement *domain.Announcement) error
	DeleteAnnouncement(ctx context.Context, id uuid.UUID) error
}

type EventMembershipRepositoryI interface {
	IsInternalMember(ctx context.Context, userID uuid.UUID, eventID uuid.UUID) (bool, error)
}

type AnnouncementUseCase struct {
	repo       AnnouncementRepositoryI
	membership EventMembershipRepositoryI
}

func NewAnnouncementUseCase(repo AnnouncementRepositoryI, membership EventMembershipRepositoryI) AnnouncementUseCase {
	return AnnouncementUseCase{
		repo:       repo,
		membership: membership,
	}
}

type UpdateAnnouncementInput struct {
	Body string
}

func (uc AnnouncementUseCase) GetAnnouncement(ctx context.Context, page shared.Pagination) ([]domain.Announcement, error) {
	return uc.repo.GetAnnouncement(ctx, page)
}

func (uc AnnouncementUseCase) GetAnnouncementById(ctx context.Context, id uuid.UUID) (*domain.Announcement, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc AnnouncementUseCase) UpdateAnnouncement(ctx context.Context, userID uuid.UUID, id uuid.UUID, input UpdateAnnouncementInput) (*domain.Announcement, error) {
	announcement, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := uc.checkEventAccess(ctx, userID, announcement.EventId); err != nil {
		return nil, err
	}

	announcement.Body = input.Body

	if err := uc.repo.UpdateAnnouncement(ctx, announcement); err != nil {
		return nil, err
	}

	return announcement, nil
}

func (uc AnnouncementUseCase) DeleteAnnouncement(ctx context.Context, userID uuid.UUID, id uuid.UUID) error {
	announcement, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := uc.checkEventAccess(ctx, userID, announcement.EventId); err != nil {
		return err
	}

	return uc.repo.DeleteAnnouncement(ctx, id)
}

func (uc AnnouncementUseCase) checkEventAccess(
	ctx context.Context,
	userID uuid.UUID,
	eventID uuid.UUID,
) error {
	isMember, err := uc.membership.IsInternalMember(ctx, userID, eventID)
	if err != nil {
		return err
	}

	if !isMember {
		return domain.ErrAnnouncementForbidden
	}

	return nil
}
