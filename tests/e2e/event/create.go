//go:build e2e

package event_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"main/internal/infrastructure/database"
	"main/internal/modules/event/domain"
	"main/internal/modules/event/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateEvent(t *testing.T) {
	a := newAPI(t)
	token := a.newUserToken(t)

	t.Run("creates an upcoming event with the whole quota available", func(t *testing.T) {
		req := newCreateRequest()

		w := a.do(t, http.MethodPost, eventsPath, req, token)

		require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
		got := decodeData[dto.EventResponse](t, w)
		_, err := uuid.Parse(got.ID)
		assert.NoError(t, err, "the id is a uuid")
		assert.Equal(t, req.Title, got.Title)
		assert.Equal(t, req.Description, got.Description)
		assert.Equal(t, req.Location, got.Location)
		assert.True(t, req.Date.Equal(got.Date), "date %v, want %v", got.Date, req.Date)
		assert.Equal(t, req.MaxQuota, got.MaxQuota)
		assert.Equal(t, req.MaxQuota, got.AvailableQuota)
		assert.Equal(t, string(domain.StatusUpcoming), got.Status)
	})

	t.Run("the creator becomes its coordinator", func(t *testing.T) {
		event := createEvent(t, a, token, newCreateRequest())

		var members []database.Committee
		require.NoError(t, a.db.Where("event_id = ?", event.ID).Find(&members).Error)
		require.Len(t, members, 1)
		assert.Equal(t, database.RoleCoordinator, members[0].Role)

		w := a.do(t, http.MethodGet, myEventsPath, nil, token)
		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Contains(t, eventIDs(decodeList(t, w).Data), event.ID)
	})

	t.Run("the creator can edit it right away", func(t *testing.T) {
		event := createEvent(t, a, token, newCreateRequest())

		w := a.do(t, http.MethodPut, eventPath(event.ID), newUpdateRequest(), token)

		assert.Equal(t, http.StatusOK, w.Code, w.Body.String())
	})

	t.Run("rejects a date in the past", func(t *testing.T) {
		req := newCreateRequest()
		req.Date = time.Now().Add(-time.Hour)

		w := a.do(t, http.MethodPost, eventsPath, req, token)

		assert.Equal(t, http.StatusBadRequest, w.Code, w.Body.String())
	})

	t.Run("rejects invalid input", func(t *testing.T) {
		valid := func() map[string]any {
			return map[string]any{
				"title":       "Seminar Karier",
				"description": "Praktisi industri berbagi cerita.",
				"date":        futureDate(14),
				"location":    "Auditorium Gedung A",
				"max_quota":   100,
			}
		}

		cases := []struct {
			name   string
			mutate func(map[string]any)
		}{
			{"missing title", func(m map[string]any) { delete(m, "title") }},
			{"empty title", func(m map[string]any) { m["title"] = "" }},
			{"title over 200 characters", func(m map[string]any) { m["title"] = strings.Repeat("a", 201) }},
			{"missing description", func(m map[string]any) { delete(m, "description") }},
			{"missing date", func(m map[string]any) { delete(m, "date") }},
			{"date that is not a date", func(m map[string]any) { m["date"] = "tomorrow" }},
			{"missing location", func(m map[string]any) { delete(m, "location") }},
			{"location over 255 characters", func(m map[string]any) { m["location"] = strings.Repeat("a", 256) }},
			{"missing max quota", func(m map[string]any) { delete(m, "max_quota") }},
			{"zero max quota", func(m map[string]any) { m["max_quota"] = 0 }},
			{"negative max quota", func(m map[string]any) { m["max_quota"] = -5 }},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				body := valid()
				tc.mutate(body)

				w := a.do(t, http.MethodPost, eventsPath, body, token)

				assertInvalid(t, w)
			})
		}
	})

	t.Run("requires a signed-in user", func(t *testing.T) {
		w := a.do(t, http.MethodPost, eventsPath, newCreateRequest(), "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
