package usecase

import (
	"context"
	"main/internal/modules/registration/domain"
)

type registrationUsecase struct {
	repo  domain.Repository
	authz domain.CommitteeAuthorizer
}

func NewRegistrationUsecase(repo domain.Repository, authz domain.CommitteeAuthorizer) domain.Usecase {
	return &registrationUsecase{
		repo:  repo,
		authz: authz,
	}
}

func normalizePagination(page, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}

func (u *registrationUsecase) RegisterEvent(ctx context.Context, eventID, userID string) (*domain.Registration, error) {
	return u.repo.CreateWithQuotaDecrement(ctx, eventID, userID)
}

func (u *registrationUsecase) GetMyRegistrations(ctx context.Context, userID string, page, limit int) ([]domain.Registration, int64, error) {
	page, limit = normalizePagination(page, limit)
	return u.repo.List(ctx, domain.ListFilter{
		UserID: userID,
		Page:   page,
		Limit:  limit,
	})
}

func (u *registrationUsecase) GetRegistrations(ctx context.Context, eventID string, page, limit int) ([]domain.Registration, int64, error) {
	page, limit = normalizePagination(page, limit)
	return u.repo.List(ctx, domain.ListFilter{
		EventID: eventID,
		Page:    page,
		Limit:   limit,
	})
}

func (u *registrationUsecase) CancelRegistration(ctx context.Context, registrationID, actorUserID string) (*domain.Registration, error) {
	reg, err := u.repo.FindByID(ctx, registrationID)
	if err != nil {
		return nil, err
	}

	if reg.UserID != actorUserID {
		isCoord, err := u.authz.IsCoordinator(ctx, reg.EventID, actorUserID)
		if err != nil {
			return nil, err
		}
		if !isCoord {
			return nil, domain.ErrForbidden
		}
	}

	return u.repo.CancelWithQuotaIncrement(ctx, registrationID)
}

func (u *registrationUsecase) UpdateRegistration(ctx context.Context, registrationID, actorUserID string, status domain.Status) (*domain.Registration, error) {
	if !status.IsValid() {
		return nil, domain.ErrInvalidStatus
	}

	reg, err := u.repo.FindByID(ctx, registrationID)
	if err != nil {
		return nil, err
	}

	isCoord, err := u.authz.IsCoordinator(ctx, reg.EventID, actorUserID)
	if err != nil {
		return nil, err
	}
	if !isCoord {
		return nil, domain.ErrForbidden
	}

	return u.repo.UpdateStatusWithQuotaSync(ctx, registrationID, status)
}

func (u *registrationUsecase) DeleteRegistration(ctx context.Context, registrationID, actorUserID string) error {
	reg, err := u.repo.FindByID(ctx, registrationID)
	if err != nil {
		return err
	}

	isCoord, err := u.authz.IsCoordinator(ctx, reg.EventID, actorUserID)
	if err != nil {
		return err
	}
	if !isCoord {
		return domain.ErrForbidden
	}

	return u.repo.DeleteWithQuotaSync(ctx, registrationID)
}

func (u *registrationUsecase) CheckInRegistration(ctx context.Context, registrationID, actorUserID string) (*domain.Registration, error) {
	reg, err := u.repo.FindByID(ctx, registrationID)
	if err != nil {
		return nil, err
	}

	isMember, err := u.authz.IsCommitteeMember(ctx, reg.EventID, actorUserID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, domain.ErrForbidden
	}

	switch reg.Status {
	case domain.StatusCheckedIn:
		return nil, domain.ErrAlreadyCheckedIn
	case domain.StatusCancelled:
		return nil, domain.ErrCannotCheckInCancelled
	}

	return u.repo.UpdateStatusWithQuotaSync(ctx, registrationID, domain.StatusCheckedIn)
}
