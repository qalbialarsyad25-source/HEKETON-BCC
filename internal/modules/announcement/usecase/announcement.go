
package usecase

import (
	"context"
	"encoding/json"

	"main/internal/modules/announcement/domain"
	shared "main/internal/shared/pagination"

	"github.com/google/uuid"
)

type AnnouncementRepositoryI interface {
	CreateAnnouncement(ctx context.Context, announcement *domain.Announcement) error
	GetAnnouncement(ctx context.Context, page shared.Pagination) ([]domain.Announcement, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Announcement, error)
	UpdateAnnouncement(ctx context.Context, announcement *domain.Announcement) error
	DeleteAnnouncement(ctx context.Context, id uuid.UUID) error
}

type EventMembershipRepositoryI interface {
	IsInternalMember(ctx context.Context, userID uuid.UUID, eventID uuid.UUID) (bool, error)
}

type AnnouncementPublisher interface {
	Publish(eventID uuid.UUID, message []byte)
}

type AnnouncementUseCase struct {
	repo       AnnouncementRepositoryI
	membership EventMembershipRepositoryI
	publisher  AnnouncementPublisher
}

func NewAnnouncementUseCase(
	repo AnnouncementRepositoryI,
	membership EventMembershipRepositoryI,
	publisher AnnouncementPublisher,
) AnnouncementUseCase {
	return AnnouncementUseCase{
		repo:       repo,
		membership: membership,
		publisher:  publisher,
	}
}

type UpdateAnnouncementInput struct {
	Body string
}

func (uc AnnouncementUseCase) CreateAnnouncement(
	ctx context.Context,
	userID uuid.UUID,
	title string,
	body string,
	eventID uuid.UUID,
) (*domain.Announcement, error) {

	if err := uc.checkEventAccess(ctx, userID, eventID); err != nil {
		return nil, err
	}

	announcement := &domain.Announcement{
		ID:       uuid.New(),
		Title:    title,
		Body:     body,
		SenderId: userID,
		EventId:  eventID,
	}

	if err := uc.repo.CreateAnnouncement(ctx, announcement); err != nil {
		return nil, err
	}

	uc.publishAnnouncement(
		announcement.EventId,
		"announcement.created",
		announcement,
	)

	return announcement, nil
}

func (uc AnnouncementUseCase) GetAnnouncement(
	ctx context.Context,
	page shared.Pagination,
) ([]domain.Announcement, error) {
	return uc.repo.GetAnnouncement(ctx, page)
}

func (uc AnnouncementUseCase) GetAnnouncementById(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Announcement, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc AnnouncementUseCase) UpdateAnnouncement(
	ctx context.Context,
	userID uuid.UUID,
	id uuid.UUID,
	input UpdateAnnouncementInput,
) (*domain.Announcement, error) {

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

	uc.publishAnnouncement(
		announcement.EventId,
		"announcement.updated",
		announcement,
	)

	return announcement, nil
}

func (uc AnnouncementUseCase) DeleteAnnouncement(
	ctx context.Context,
	userID uuid.UUID,
	id uuid.UUID,
) error {

	announcement, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := uc.checkEventAccess(ctx, userID, announcement.EventId); err != nil {
		return err
	}

	if err := uc.repo.DeleteAnnouncement(ctx, id); err != nil {
		return err
	}

	uc.publishAnnouncement(
		announcement.EventId,
		"announcement.deleted",
		map[string]any{
			"id":      announcement.ID,
			"eventId": announcement.EventId,
		},
	)

	return nil
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

func (uc AnnouncementUseCase) publishAnnouncement(
	eventID uuid.UUID,
	eventType string,
	data any,
) {
	if uc.publisher == nil {
		return
	}

	message, err := json.Marshal(map[string]any{
		"type": eventType,
		"data": data,
	})

	if err != nil {
		return
	}

	uc.publisher.Publish(eventID, message)
}
