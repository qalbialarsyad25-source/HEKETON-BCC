
package usecase

import (
	"context"
	"errors"
	"testing"

	"main/internal/modules/announcement/domain"
	shared "main/internal/shared/pagination"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAnnouncementRepository struct {
	announcement  *domain.Announcement
	announcements []domain.Announcement
	err           error
	updateCalled  bool
	deleteCalled  bool
}

func (m *mockAnnouncementRepository) GetAnnouncement(
	ctx context.Context,
	page shared.Pagination,
) ([]domain.Announcement, error) {
	return m.announcements, m.err
}

func (m *mockAnnouncementRepository) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Announcement, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.announcement, nil
}

func (m *mockAnnouncementRepository) UpdateAnnouncement(
	ctx context.Context,
	announcement *domain.Announcement,
) error {
	m.updateCalled = true
	return m.err
}

func (m *mockAnnouncementRepository) DeleteAnnouncement(
	ctx context.Context,
	id uuid.UUID,
) error {
	m.deleteCalled = true
	return m.err
}

type mockEventMembershipRepository struct {
	isMember bool
	err      error
}

func (m *mockEventMembershipRepository) IsInternalMember(
	ctx context.Context,
	userID uuid.UUID,
	eventID uuid.UUID,
) (bool, error) {
	return m.isMember, m.err
}

func newTestAnnouncement() *domain.Announcement {
	return &domain.Announcement{
		ID:       uuid.New(),
		Title:    "Announcement Test",
		Body:     "Announcement body",
		SenderId: uuid.New(),
		EventId:  uuid.New(),
	}
}

func TestGetAnnouncement(t *testing.T) {
	t.Run("success get announcements", func(t *testing.T) {
		announcements := []domain.Announcement{
			*newTestAnnouncement(),
			*newTestAnnouncement(),
		}

		repo := &mockAnnouncementRepository{
			announcements: announcements,
		}

		membership := &mockEventMembershipRepository{}

		uc := NewAnnouncementUseCase(repo, membership)

		result, err := uc.GetAnnouncement(
			context.Background(),
			shared.Pagination{},
		)

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, announcements, result)
	})

	t.Run("error when failed to get announcements", func(t *testing.T) {
		expectedErr := errors.New("failed to get announcements")

		repo := &mockAnnouncementRepository{
			err: expectedErr,
		}

		membership := &mockEventMembershipRepository{}

		uc := NewAnnouncementUseCase(repo, membership)

		result, err := uc.GetAnnouncement(
			context.Background(),
			shared.Pagination{},
		)

		require.ErrorIs(t, err, expectedErr)
		assert.Nil(t, result)
	})
}

func TestGetAnnouncementById(t *testing.T) {
	t.Run("success get announcement by id", func(t *testing.T) {
		announcement := newTestAnnouncement()

		repo := &mockAnnouncementRepository{
			announcement: announcement,
		}

		membership := &mockEventMembershipRepository{}

		uc := NewAnnouncementUseCase(repo, membership)

		result, err := uc.GetAnnouncementById(
			context.Background(),
			announcement.ID,
		)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, announcement.ID, result.ID)
		assert.Equal(t, announcement.Title, result.Title)
	})

	t.Run("error when announcement not found", func(t *testing.T) {
		expectedErr := errors.New("announcement not found")

		repo := &mockAnnouncementRepository{
			err: expectedErr,
		}

		membership := &mockEventMembershipRepository{}

		uc := NewAnnouncementUseCase(repo, membership)

		result, err := uc.GetAnnouncementById(
			context.Background(),
			uuid.New(),
		)

		require.ErrorIs(t, err, expectedErr)
		assert.Nil(t, result)
	})
}

func TestUpdateAnnouncement(t *testing.T) {
	t.Run("success update announcement", func(t *testing.T) {
		announcement := newTestAnnouncement()

		repo := &mockAnnouncementRepository{
			announcement: announcement,
		}

		membership := &mockEventMembershipRepository{
			isMember: true,
		}

		uc := NewAnnouncementUseCase(repo, membership)

		result, err := uc.UpdateAnnouncement(
			context.Background(),
			uuid.New(),
			announcement.ID,
			UpdateAnnouncementInput{
				Body: "Updated announcement body",
			},
		)

		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "Updated announcement body", result.Body)
		assert.True(t, repo.updateCalled)
	})

	t.Run("forbidden when user is not internal member", func(t *testing.T) {
		announcement := newTestAnnouncement()

		repo := &mockAnnouncementRepository{
			announcement: announcement,
		}

		membership := &mockEventMembershipRepository{
			isMember: false,
		}

		uc := NewAnnouncementUseCase(repo, membership)

		result, err := uc.UpdateAnnouncement(
			context.Background(),
			uuid.New(),
			announcement.ID,
			UpdateAnnouncementInput{
				Body: "Updated announcement body",
			},
		)

		require.ErrorIs(t, err, domain.ErrAnnouncementForbidden)
		assert.Nil(t, result)
		assert.False(t, repo.updateCalled)
	})

	t.Run("error when announcement not found", func(t *testing.T) {
		expectedErr := errors.New("announcement not found")

		repo := &mockAnnouncementRepository{
			err: expectedErr,
		}

		membership := &mockEventMembershipRepository{
			isMember: true,
		}

		uc := NewAnnouncementUseCase(repo, membership)

		result, err := uc.UpdateAnnouncement(
			context.Background(),
			uuid.New(),
			uuid.New(),
			UpdateAnnouncementInput{
				Body: "Updated announcement body",
			},
		)

		require.ErrorIs(t, err, expectedErr)
		assert.Nil(t, result)
		assert.False(t, repo.updateCalled)
	})
}

func TestDeleteAnnouncement(t *testing.T) {
	t.Run("success delete announcement", func(t *testing.T) {
		announcement := newTestAnnouncement()

		repo := &mockAnnouncementRepository{
			announcement: announcement,
		}

		membership := &mockEventMembershipRepository{
			isMember: true,
		}

		uc := NewAnnouncementUseCase(repo, membership)

		err := uc.DeleteAnnouncement(
			context.Background(),
			uuid.New(),
			announcement.ID,
		)

		require.NoError(t, err)
		assert.True(t, repo.deleteCalled)
	})

	t.Run("forbidden when user is not internal member", func(t *testing.T) {
		announcement := newTestAnnouncement()

		repo := &mockAnnouncementRepository{
			announcement: announcement,
		}

		membership := &mockEventMembershipRepository{
			isMember: false,
		}

		uc := NewAnnouncementUseCase(repo, membership)

		err := uc.DeleteAnnouncement(
			context.Background(),
			uuid.New(),
			announcement.ID,
		)

		require.ErrorIs(t, err, domain.ErrAnnouncementForbidden)
		assert.False(t, repo.deleteCalled)
	})

	t.Run("error when announcement not found", func(t *testing.T) {
		expectedErr := errors.New("announcement not found")

		repo := &mockAnnouncementRepository{
			err: expectedErr,
		}

		membership := &mockEventMembershipRepository{
			isMember: true,
		}

		uc := NewAnnouncementUseCase(repo, membership)

		err := uc.DeleteAnnouncement(
			context.Background(),
			uuid.New(),
			uuid.New(),
		)

		require.ErrorIs(t, err, expectedErr)
		assert.False(t, repo.deleteCalled)
	})
}
