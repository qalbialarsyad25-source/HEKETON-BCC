package usecase

import (
	"context"
	"testing"
	"time"

	"main/internal/modules/session/domain"
	"main/internal/modules/session/factory"
	"main/internal/shared/pagination"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSessionRepository struct {
	session     *domain.Session
	sessions    []domain.Session
	total       int64
	eventExists bool
	isMember    bool
	updateGone  bool
	deleteGone  bool
	err         error

	createCalled bool
	updateCalled bool
	deleteCalled bool
	listEventID  *uuid.UUID
}

func (m *mockSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	m.createCalled = true
	if m.err != nil {
		return m.err
	}
	m.session = session
	return nil
}

func (m *mockSessionRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.session, nil
}

func (m *mockSessionRepository) List(ctx context.Context, eventID *uuid.UUID, params pagination.Params) ([]domain.Session, int64, error) {
	m.listEventID = eventID
	return m.sessions, m.total, m.err
}

func (m *mockSessionRepository) Update(ctx context.Context, session *domain.Session) (*domain.Session, error) {
	m.updateCalled = true
	if m.err != nil {
		return nil, m.err
	}
	if m.updateGone {
		return nil, nil
	}
	return session, nil
}

func (m *mockSessionRepository) Delete(ctx context.Context, id uuid.UUID) (bool, error) {
	m.deleteCalled = true
	if m.err != nil {
		return false, m.err
	}
	return !m.deleteGone, nil
}

func (m *mockSessionRepository) EventExists(ctx context.Context, eventID uuid.UUID) (bool, error) {
	return m.eventExists, m.err
}

func (m *mockSessionRepository) IsCommitteeMember(ctx context.Context, eventID, userID uuid.UUID) (bool, error) {
	return m.isMember, m.err
}

type mockNotifier struct {
	created []domain.Session
	updated []domain.Session
	deleted []domain.Session
}

func (m *mockNotifier) SessionCreated(session domain.Session) { m.created = append(m.created, session) }
func (m *mockNotifier) SessionUpdated(session domain.Session) { m.updated = append(m.updated, session) }
func (m *mockNotifier) SessionDeleted(session domain.Session) { m.deleted = append(m.deleted, session) }

func TestCreateSession(t *testing.T) {
	input := CreateSessionInput{
		UserID:   uuid.New(),
		EventID:  uuid.New(),
		Title:    "Keynote",
		Time:     time.Now().Add(time.Hour),
		Speaker:  "Speaker",
		Location: "Auditorium",
	}

	t.Run("a committee member creates an upcoming session and followers are told", func(t *testing.T) {
		repo := &mockSessionRepository{eventExists: true, isMember: true}
		notifier := &mockNotifier{}
		uc := NewSessionUseCase(repo, notifier)

		result, err := uc.Create(context.Background(), input)

		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, repo.createCalled)
		assert.NotEqual(t, uuid.Nil, result.ID)
		assert.Equal(t, input.EventID, result.EventID)
		assert.Equal(t, input.Title, result.Title)
		assert.Equal(t, input.Speaker, result.Speaker)
		assert.Equal(t, input.Location, result.Location)
		assert.Equal(t, domain.StatusUpcoming, result.Status)
		require.Len(t, notifier.created, 1)
		assert.Equal(t, result.ID, notifier.created[0].ID)
	})

	t.Run("works without a notifier", func(t *testing.T) {
		repo := &mockSessionRepository{eventExists: true, isMember: true}
		uc := NewSessionUseCase(repo, nil)

		result, err := uc.Create(context.Background(), input)

		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("unknown event is not found", func(t *testing.T) {
		repo := &mockSessionRepository{eventExists: false, isMember: true}
		notifier := &mockNotifier{}
		uc := NewSessionUseCase(repo, notifier)

		result, err := uc.Create(context.Background(), input)

		require.ErrorIs(t, err, domain.ErrEventNotFound)
		assert.Nil(t, result)
		assert.False(t, repo.createCalled)
		assert.Empty(t, notifier.created)
	})

	t.Run("someone outside the committee is forbidden", func(t *testing.T) {
		repo := &mockSessionRepository{eventExists: true, isMember: false}
		notifier := &mockNotifier{}
		uc := NewSessionUseCase(repo, notifier)

		result, err := uc.Create(context.Background(), input)

		require.ErrorIs(t, err, domain.ErrNotEventMember)
		assert.Nil(t, result)
		assert.False(t, repo.createCalled)
		assert.Empty(t, notifier.created)
	})
}

func TestGetSession(t *testing.T) {
	t.Run("returns the session", func(t *testing.T) {
		session := factory.NewSession(uuid.New())
		uc := NewSessionUseCase(&mockSessionRepository{session: session}, nil)

		result, err := uc.GetByID(context.Background(), session.ID)

		require.NoError(t, err)
		assert.Equal(t, session.ID, result.ID)
	})

	t.Run("unknown id is not found", func(t *testing.T) {
		uc := NewSessionUseCase(&mockSessionRepository{}, nil)

		result, err := uc.GetByID(context.Background(), uuid.New())

		require.ErrorIs(t, err, domain.ErrSessionNotFound)
		assert.Nil(t, result)
	})
}

func TestListSessions(t *testing.T) {
	sessions := []domain.Session{*factory.NewSession(uuid.New()), *factory.NewSession(uuid.New())}
	repo := &mockSessionRepository{sessions: sessions, total: 25}
	uc := NewSessionUseCase(repo, nil)
	eventID := uuid.New()

	result, err := uc.List(context.Background(), &eventID, pagination.NewParams(2, 10))

	require.NoError(t, err)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, pagination.Meta{Page: 2, Limit: 10, TotalItems: 25, TotalPages: 3}, result.Meta)
	require.NotNil(t, repo.listEventID)
	assert.Equal(t, eventID, *repo.listEventID)
}

func TestUpdateSession(t *testing.T) {
	newInput := func(session *domain.Session) UpdateSessionInput {
		return UpdateSessionInput{
			ID:       session.ID,
			UserID:   uuid.New(),
			Title:    "New title",
			Time:     time.Now().Add(48 * time.Hour),
			Speaker:  "New speaker",
			Location: "New location",
		}
	}

	t.Run("a committee member updates the fields and keeps the status", func(t *testing.T) {
		session := factory.NewSession(uuid.New())
		session.Status = domain.StatusOngoing
		repo := &mockSessionRepository{session: session, isMember: true}
		notifier := &mockNotifier{}
		uc := NewSessionUseCase(repo, notifier)
		input := newInput(session)

		result, err := uc.Update(context.Background(), input)

		require.NoError(t, err)
		assert.True(t, repo.updateCalled)
		assert.Equal(t, input.Title, result.Title)
		assert.Equal(t, input.Speaker, result.Speaker)
		assert.Equal(t, input.Location, result.Location)
		assert.True(t, input.Time.Equal(result.Time))
		assert.Equal(t, domain.StatusOngoing, result.Status)
		require.Len(t, notifier.updated, 1)
		assert.Equal(t, session.ID, notifier.updated[0].ID)
	})

	t.Run("changes the status when it is given", func(t *testing.T) {
		session := factory.NewSession(uuid.New())
		repo := &mockSessionRepository{session: session, isMember: true}
		uc := NewSessionUseCase(repo, nil)
		input := newInput(session)
		input.Status = domain.StatusFinished

		result, err := uc.Update(context.Background(), input)

		require.NoError(t, err)
		assert.Equal(t, domain.StatusFinished, result.Status)
	})

	t.Run("rejects an unknown status", func(t *testing.T) {
		session := factory.NewSession(uuid.New())
		repo := &mockSessionRepository{session: session, isMember: true}
		uc := NewSessionUseCase(repo, nil)
		input := newInput(session)
		input.Status = domain.Status("CANCELLED")

		result, err := uc.Update(context.Background(), input)

		require.ErrorIs(t, err, domain.ErrInvalidStatus)
		assert.Nil(t, result)
		assert.False(t, repo.updateCalled)
	})

	t.Run("someone outside the committee is forbidden", func(t *testing.T) {
		session := factory.NewSession(uuid.New())
		repo := &mockSessionRepository{session: session, isMember: false}
		notifier := &mockNotifier{}
		uc := NewSessionUseCase(repo, notifier)

		result, err := uc.Update(context.Background(), newInput(session))

		require.ErrorIs(t, err, domain.ErrNotEventMember)
		assert.Nil(t, result)
		assert.False(t, repo.updateCalled)
		assert.Empty(t, notifier.updated)
	})

	t.Run("unknown id is not found", func(t *testing.T) {
		repo := &mockSessionRepository{isMember: true}
		uc := NewSessionUseCase(repo, nil)

		result, err := uc.Update(context.Background(), newInput(factory.NewSession(uuid.New())))

		require.ErrorIs(t, err, domain.ErrSessionNotFound)
		assert.Nil(t, result)
		assert.False(t, repo.updateCalled)
	})

	t.Run("a session deleted meanwhile is not found", func(t *testing.T) {
		session := factory.NewSession(uuid.New())
		repo := &mockSessionRepository{session: session, isMember: true, updateGone: true}
		notifier := &mockNotifier{}
		uc := NewSessionUseCase(repo, notifier)

		result, err := uc.Update(context.Background(), newInput(session))

		require.ErrorIs(t, err, domain.ErrSessionNotFound)
		assert.Nil(t, result)
		assert.Empty(t, notifier.updated)
	})
}

func TestDeleteSession(t *testing.T) {
	t.Run("a committee member deletes the session and followers are told", func(t *testing.T) {
		session := factory.NewSession(uuid.New())
		repo := &mockSessionRepository{session: session, isMember: true}
		notifier := &mockNotifier{}
		uc := NewSessionUseCase(repo, notifier)

		result, err := uc.Delete(context.Background(), session.ID, uuid.New())

		require.NoError(t, err)
		assert.Equal(t, session.ID, result.ID)
		assert.True(t, repo.deleteCalled)
		require.Len(t, notifier.deleted, 1)
		assert.Equal(t, session.ID, notifier.deleted[0].ID)
		assert.Equal(t, session.EventID, notifier.deleted[0].EventID)
	})

	t.Run("someone outside the committee is forbidden", func(t *testing.T) {
		session := factory.NewSession(uuid.New())
		repo := &mockSessionRepository{session: session, isMember: false}
		notifier := &mockNotifier{}
		uc := NewSessionUseCase(repo, notifier)

		result, err := uc.Delete(context.Background(), session.ID, uuid.New())

		require.ErrorIs(t, err, domain.ErrNotEventMember)
		assert.Nil(t, result)
		assert.False(t, repo.deleteCalled)
		assert.Empty(t, notifier.deleted)
	})

	t.Run("unknown id is not found", func(t *testing.T) {
		repo := &mockSessionRepository{isMember: true}
		uc := NewSessionUseCase(repo, nil)

		result, err := uc.Delete(context.Background(), uuid.New(), uuid.New())

		require.ErrorIs(t, err, domain.ErrSessionNotFound)
		assert.Nil(t, result)
		assert.False(t, repo.deleteCalled)
	})

	t.Run("a session deleted meanwhile is not found", func(t *testing.T) {
		session := factory.NewSession(uuid.New())
		repo := &mockSessionRepository{session: session, isMember: true, deleteGone: true}
		notifier := &mockNotifier{}
		uc := NewSessionUseCase(repo, notifier)

		result, err := uc.Delete(context.Background(), session.ID, uuid.New())

		require.ErrorIs(t, err, domain.ErrSessionNotFound)
		assert.Nil(t, result)
		assert.Empty(t, notifier.deleted)
	})
}
