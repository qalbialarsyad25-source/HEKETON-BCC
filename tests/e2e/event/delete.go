package event_test

import (
	"net/http"
	"testing"

	"main/internal/infrastructure/database"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeleteEvent(t *testing.T) {
	a := newAPI(t)
	coordinator := a.newUserToken(t)
	stranger := a.newUserToken(t)

	t.Run("a coordinator deletes the event and its committee", func(t *testing.T) {
		event := createEvent(t, a, coordinator, newCreateRequest())
		var before int64
		require.NoError(t, a.db.Model(&database.Committee{}).Where("event_id = ?", event.ID).Count(&before).Error)
		require.Equal(t, int64(1), before)

		w := a.do(t, http.MethodDelete, eventPath(event.ID), nil, coordinator)

		assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
		read := a.do(t, http.MethodGet, eventPath(event.ID), nil, coordinator)
		assert.Equal(t, http.StatusNotFound, read.Code, "the event is gone")
		var after int64
		require.NoError(t, a.db.Model(&database.Committee{}).Where("event_id = ?", event.ID).Count(&after).Error)
		assert.Equal(t, int64(0), after, "ON DELETE CASCADE removes the committee")
	})

	t.Run("the deleted event leaves the user's list", func(t *testing.T) {
		token := a.newUserToken(t)
		event := createEvent(t, a, token, newCreateRequest())

		w := a.do(t, http.MethodDelete, eventPath(event.ID), nil, token)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())

		list := decodeList(t, a.do(t, http.MethodGet, myEventsPath, nil, token))
		assert.NotContains(t, eventIDs(list.Data), event.ID)
	})

	t.Run("deleting twice is 404 the second time", func(t *testing.T) {
		event := createEvent(t, a, coordinator, newCreateRequest())

		first := a.do(t, http.MethodDelete, eventPath(event.ID), nil, coordinator)
		second := a.do(t, http.MethodDelete, eventPath(event.ID), nil, coordinator)

		assert.Equal(t, http.StatusOK, first.Code)
		assert.Equal(t, http.StatusNotFound, second.Code)
	})

	t.Run("someone who is not a coordinator gets 403 and the event stays", func(t *testing.T) {
		event := createEvent(t, a, coordinator, newCreateRequest())

		w := a.do(t, http.MethodDelete, eventPath(event.ID), nil, stranger)

		assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
		read := a.do(t, http.MethodGet, eventPath(event.ID), nil, coordinator)
		assert.Equal(t, http.StatusOK, read.Code, "the event is still there")
	})

	t.Run("an unknown id is 404", func(t *testing.T) {
		w := a.do(t, http.MethodDelete, eventPath(uuid.NewString()), nil, coordinator)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("an id that is not a uuid is 404", func(t *testing.T) {
		w := a.do(t, http.MethodDelete, eventPath("not-a-uuid"), nil, coordinator)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("requires a signed-in user", func(t *testing.T) {
		event := createEvent(t, a, coordinator, newCreateRequest())

		w := a.do(t, http.MethodDelete, eventPath(event.ID), nil, "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
