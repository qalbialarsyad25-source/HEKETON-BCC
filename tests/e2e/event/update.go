package event_test

import (
	"net/http"
	"testing"

	"main/internal/modules/event/domain"
	"main/internal/modules/event/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateEvent(t *testing.T) {
	a := newAPI(t)
	coordinator := a.newUserToken(t)
	stranger := a.newUserToken(t)

	t.Run("a coordinator updates the event and keeps its status", func(t *testing.T) {
		event := createEvent(t, a, coordinator, newCreateRequest())
		req := newUpdateRequest()

		w := a.do(t, http.MethodPut, eventPath(event.ID), req, coordinator)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		got := decodeData[dto.EventResponse](t, w)
		assert.Equal(t, event.ID, got.ID)
		assert.Equal(t, req.Title, got.Title)
		assert.Equal(t, req.Description, got.Description)
		assert.Equal(t, req.Location, got.Location)
		assert.True(t, req.Date.Equal(got.Date), "date %v, want %v", got.Date, req.Date)
		assert.Equal(t, req.MaxQuota, got.MaxQuota)
		assert.Equal(t, req.MaxQuota, got.AvailableQuota, "nobody registered, so everything is available")
		assert.Equal(t, string(domain.StatusUpcoming), got.Status)

		read := a.do(t, http.MethodGet, eventPath(event.ID), nil, coordinator)
		require.Equal(t, http.StatusOK, read.Code, read.Body.String())
		assertSameEvent(t, got, decodeData[dto.EventResponse](t, read))
	})

	t.Run("changes the status when it is given", func(t *testing.T) {
		event := createEvent(t, a, coordinator, newCreateRequest())
		req := newUpdateRequest()
		req.Status = string(domain.StatusOngoing)

		w := a.do(t, http.MethodPut, eventPath(event.ID), req, coordinator)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Equal(t, string(domain.StatusOngoing), decodeData[dto.EventResponse](t, w).Status)
	})

	t.Run("rejects an unknown status", func(t *testing.T) {
		event := createEvent(t, a, coordinator, newCreateRequest())
		req := newUpdateRequest()
		req.Status = "CANCELLED"

		w := a.do(t, http.MethodPut, eventPath(event.ID), req, coordinator)

		assertInvalid(t, w)
	})

	t.Run("the available quota moves with the max quota", func(t *testing.T) {
		event := createEventWithRegistrations(t, a, coordinator, 10, 5)
		req := newUpdateRequest()
		req.MaxQuota = 8

		w := a.do(t, http.MethodPut, eventPath(event.ID), req, coordinator)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		got := decodeData[dto.EventResponse](t, w)
		assert.Equal(t, 8, got.MaxQuota)
		assert.Equal(t, 3, got.AvailableQuota, "5 of the 8 seats are taken")
	})

	t.Run("the max quota can drop to exactly the registered count", func(t *testing.T) {
		event := createEventWithRegistrations(t, a, coordinator, 10, 5)
		req := newUpdateRequest()
		req.MaxQuota = 5

		w := a.do(t, http.MethodPut, eventPath(event.ID), req, coordinator)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		got := decodeData[dto.EventResponse](t, w)
		assert.Equal(t, 5, got.MaxQuota)
		assert.Equal(t, 0, got.AvailableQuota)
	})

	t.Run("a max quota below the registered count is a conflict and changes nothing", func(t *testing.T) {
		event := createEventWithRegistrations(t, a, coordinator, 10, 5)
		req := newUpdateRequest()
		req.MaxQuota = 4

		w := a.do(t, http.MethodPut, eventPath(event.ID), req, coordinator)

		assert.Equal(t, http.StatusConflict, w.Code, w.Body.String())
		read := a.do(t, http.MethodGet, eventPath(event.ID), nil, coordinator)
		require.Equal(t, http.StatusOK, read.Code, read.Body.String())
		got := decodeData[dto.EventResponse](t, read)
		assert.Equal(t, event.Title, got.Title)
		assert.Equal(t, 10, got.MaxQuota)
		assert.Equal(t, 5, got.AvailableQuota)
	})

	t.Run("someone who is not a coordinator gets 403 and changes nothing", func(t *testing.T) {
		event := createEvent(t, a, coordinator, newCreateRequest())

		w := a.do(t, http.MethodPut, eventPath(event.ID), newUpdateRequest(), stranger)

		assert.Equal(t, http.StatusForbidden, w.Code, w.Body.String())
		read := a.do(t, http.MethodGet, eventPath(event.ID), nil, coordinator)
		require.Equal(t, http.StatusOK, read.Code, read.Body.String())
		assert.Equal(t, event.Title, decodeData[dto.EventResponse](t, read).Title)
	})

	t.Run("an unknown id is 404", func(t *testing.T) {
		w := a.do(t, http.MethodPut, eventPath(uuid.NewString()), newUpdateRequest(), coordinator)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("an id that is not a uuid is 404", func(t *testing.T) {
		w := a.do(t, http.MethodPut, eventPath("not-a-uuid"), newUpdateRequest(), coordinator)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("rejects invalid input", func(t *testing.T) {
		event := createEvent(t, a, coordinator, newCreateRequest())
		valid := func() map[string]any {
			return map[string]any{
				"title":       "Judul Baru",
				"description": "Deskripsi baru.",
				"date":        futureDate(21),
				"location":    "Ruang Baru",
				"max_quota":   40,
			}
		}

		cases := []struct {
			name   string
			mutate func(map[string]any)
		}{
			{"missing title", func(m map[string]any) { delete(m, "title") }},
			{"missing description", func(m map[string]any) { delete(m, "description") }},
			{"missing date", func(m map[string]any) { delete(m, "date") }},
			{"missing location", func(m map[string]any) { delete(m, "location") }},
			{"zero max quota", func(m map[string]any) { m["max_quota"] = 0 }},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				body := valid()
				tc.mutate(body)

				w := a.do(t, http.MethodPut, eventPath(event.ID), body, coordinator)

				assertInvalid(t, w)
			})
		}
	})

	t.Run("requires a signed-in user", func(t *testing.T) {
		event := createEvent(t, a, coordinator, newCreateRequest())

		w := a.do(t, http.MethodPut, eventPath(event.ID), newUpdateRequest(), "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
