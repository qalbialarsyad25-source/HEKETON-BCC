package event_test

import (
	"net/http"
	"testing"

	"main/internal/modules/event/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEvent(t *testing.T) {
	a := newAPI(t)
	organizer := a.newUserToken(t)
	viewer := a.newUserToken(t)
	event := createEvent(t, a, organizer, newCreateRequest())

	t.Run("any signed-in user can read an event", func(t *testing.T) {
		w := a.do(t, http.MethodGet, eventPath(event.ID), nil, viewer)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assertSameEvent(t, event, decodeData[dto.EventResponse](t, w))
	})

	t.Run("an unknown id is 404", func(t *testing.T) {
		w := a.do(t, http.MethodGet, eventPath(uuid.NewString()), nil, viewer)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("an id that is not a uuid is 404", func(t *testing.T) {
		w := a.do(t, http.MethodGet, eventPath("not-a-uuid"), nil, viewer)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("requires a signed-in user", func(t *testing.T) {
		w := a.do(t, http.MethodGet, eventPath(event.ID), nil, "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
