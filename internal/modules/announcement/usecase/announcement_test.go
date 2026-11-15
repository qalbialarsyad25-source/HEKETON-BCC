package usecase

import (
	"context"
	"encoding/json"
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
	createCalled  bool
	updateCalled  bool
	deleteCalled  bool
}

func (m *mockAnnouncementRepository) CreateAnnouncement(ctx context.Context, announcement *domain.Announcement) error {
	m.createCalled = true
	if m.err != nil {
		return m.err
	}
	m.announcement = announcement
	return nil
}

func (m *mockAnnouncementRepository) GetAnnouncement(ctx context.Context, page shared.Pagination) ([]domain.Announcement, error) {
	return m.announcements, m.err
}

func (m *mockAnnouncementRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Announcement, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.announcement, nil
}

func (m *mockAnnouncementRepository) UpdateAnnouncement(ctx context.Context, announcement *domain.Announcement) error {
	m.updateCalled = true
	return m.err
}

func (m *mockAnnouncementRepository) DeleteAnnouncement(ctx context.Context, id uuid.UUID) error {
	m.deleteCalled = true
	return m.err
}

type mockEventMembershipRepository struct {
	isMember bool
	err      error
}

func (m *mockEventMembershipRepository) IsInternalMember(ctx context.Context, userID uuid.UUID, eventID uuid.UUID) (bool, error) {
	return m.isMember, m.err
}

type mockAnnouncementPublisher struct {
	publishCalled bool
	lastEventID   uuid.UUID
	lastMessage   []byte
}

func (m *mockAnnouncementPublisher) Publish(eventID uuid.UUID, message []byte) {
	m.publishCalled = true
	m.lastEventID = eventID
	m.lastMessage = message
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

func TestCreateAnnouncement(t *testing.T) {
	t.Run("success create announcement and publish ws event", func(t *testing.T) {
		repo := &mockAnnouncementRepository{}
		membership := &mockEventMembershipRepository{isMember: true}
		publisher := &mockAnnouncementPublisher{}

		uc := NewAnnouncementUseCase(repo, membership, publisher)

		userID := uuid.New()
		eventID := uuid.New()

		result, err := uc.CreateAnnouncement(context.Background(), userID, "Judul", "Isi pengumuman", eventID)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, repo.createCalled)
		assert.True(t, publisher.publishCalled)
		assert.Equal(t, eventID, publisher.lastEventID)

		var payload map[string]any
		require.NoError(t, json.Unmarshal(publisher.lastMessage, &payload))
		assert.Equal(t, "announcement.created", payload["type"])
	})

	t.Run("forbidden when user is not internal member", func(t *testing.T) {
		repo := &mockAnnouncementRepository{}
		membership := &mockEventMembershipRepository{isMember: false}
		publisher := &mockAnnouncementPublisher{}

		uc := NewAnnouncementUseCase(repo, membership, publisher)

		result, err := uc.CreateAnnouncement(context.Background(), uuid.New(), "Judul", "Isi", uuid.New())

		require.ErrorIs(t, err, domain.ErrAnnouncementForbidden)
		assert.Nil(t, result)
		assert.False(t, publisher.publishCalled)
	})
}

func TestUpdateAnnouncement(t *testing.T) {
	t.Run("success update announcement and publish ws event", func(t *testing.T) {
		announcement := newTestAnnouncement()
		repo := &mockAnnouncementRepository{announcement: announcement}
		membership := &mockEventMembershipRepository{isMember: true}
		publisher := &mockAnnouncementPublisher{}

		uc := NewAnnouncementUseCase(repo, membership, publisher)

		result, err := uc.UpdateAnnouncement(context.Background(), uuid.New(), announcement.ID, UpdateAnnouncementInput{
			Body: "Updated body",
		})

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, repo.updateCalled)
		assert.True(t, publisher.publishCalled)
		assert.Equal(t, announcement.EventId, publisher.lastEventID)

		var payload map[string]any
		require.NoError(t, json.Unmarshal(publisher.lastMessage, &payload))
		assert.Equal(t, "announcement.updated", payload["type"])
	})
}

func TestDeleteAnnouncement(t *testing.T) {
	t.Run("success delete announcement and publish ws event", func(t *testing.T) {
		announcement := newTestAnnouncement()
		repo := &mockAnnouncementRepository{announcement: announcement}
		membership := &mockEventMembershipRepository{isMember: true}
		publisher := &mockAnnouncementPublisher{}

		uc := NewAnnouncementUseCase(repo, membership, publisher)

		err := uc.DeleteAnnouncement(context.Background(), uuid.New(), announcement.ID)

		require.NoError(t, err)
		assert.True(t, repo.deleteCalled)
		assert.True(t, publisher.publishCalled)
		assert.Equal(t, announcement.EventId, publisher.lastEventID)

		var payload map[string]any
		require.NoError(t, json.Unmarshal(publisher.lastMessage, &payload))
		assert.Equal(t, "announcement.deleted", payload["type"])
	})
}
