package event_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"main/internal/infrastructure/database"
	"main/internal/modules/event/dto"
	"main/internal/shared/authz"
	"main/tests/e2e/factory"
	"main/tests/e2e/helper"
	"main/tests/e2e/suite"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

const (
	eventsPath   = "/api/v1/events"
	myEventsPath = eventsPath + "/my"
)

func eventPath(id string) string {
	return eventsPath + "/" + id
}

// api hides the suite's router and adapters behind two functions, so every test reads the same way.
type api struct {
	db *gorm.DB
	// do sends a request. An empty token sends no Authorization header.
	do func(t *testing.T, method, path string, body any, token string) *httptest.ResponseRecorder
	// newUserToken creates a user, signs it in and returns its access token.
	newUserToken func(t *testing.T) string
}

func newAPI(t *testing.T) *api {
	t.Helper()
	ts := suite.New(t)

	return &api{
		db: ts.DB,
		do: func(t *testing.T, method, path string, body any, token string) *httptest.ResponseRecorder {
			return helper.Do(t, ts.Router, method, path, body, token)
		},
		newUserToken: func(t *testing.T) string {
			user := factory.CreateUser(t, ts.DB, ts.Adapters.Hasher, authz.RoleUser)
			session, _ := factory.Login(t, ts.Router, user.Email, factory.Password)
			return session
		},
	}
}

func futureDate(days int) time.Time {
	return time.Now().Add(time.Duration(days) * 24 * time.Hour).UTC().Truncate(time.Second)
}

func newCreateRequest() dto.CreateEventRequest {
	return dto.CreateEventRequest{
		Title:       "Workshop Backend Development dengan Golang",
		Description: "Membangun REST API dari nol dengan Go, GORM, dan autentikasi JWT.",
		Date:        futureDate(14),
		Location:    "Laboratorium Komputer Gedung F, Lantai 3",
		MaxQuota:    30,
	}
}

func newUpdateRequest() dto.UpdateEventRequest {
	return dto.UpdateEventRequest{
		Title:       "Workshop Backend Development dengan Golang (Sesi 2)",
		Description: "Sesi lanjutan: database, migrasi, dan deployment dengan Docker.",
		Date:        futureDate(21),
		Location:    "Ruang Seminar Gedung B, Lantai 4",
		MaxQuota:    40,
	}
}

func createEvent(t *testing.T, a *api, token string, req dto.CreateEventRequest) dto.EventResponse {
	t.Helper()

	w := a.do(t, http.MethodPost, eventsPath, req, token)
	require.Equal(t, http.StatusCreated, w.Code, w.Body.String())

	return decodeData[dto.EventResponse](t, w)
}

// createEventWithRegistrations creates an event and lowers its available quota as if
// `registered` participants had signed up. The registration module isn't needed for that:
// the event only stores the counter.
func createEventWithRegistrations(t *testing.T, a *api, token string, maxQuota, registered int) dto.EventResponse {
	t.Helper()

	req := newCreateRequest()
	req.MaxQuota = maxQuota
	event := createEvent(t, a, token, req)

	err := a.db.Model(&database.Event{}).
		Where("id = ?", event.ID).
		Update("available_quota", maxQuota-registered).Error
	require.NoError(t, err)

	return event
}

func decodeData[T any](t *testing.T, w *httptest.ResponseRecorder) T {
	t.Helper()

	var body struct {
		Data T `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body), w.Body.String())

	return body.Data
}

func decodeList(t *testing.T, w *httptest.ResponseRecorder) dto.EventListResponse {
	t.Helper()

	var list dto.EventListResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &list), w.Body.String())

	return list
}

func eventIDs(events []dto.EventResponse) []string {
	ids := make([]string, 0, len(events))
	for _, event := range events {
		ids = append(ids, event.ID)
	}

	return ids
}

// assertSameEvent compares field by field, because the date comes back in another time zone
// depending on whether it was just written or read from the database.
func assertSameEvent(t *testing.T, want, got dto.EventResponse) {
	t.Helper()

	assert.Equal(t, want.ID, got.ID)
	assert.Equal(t, want.Title, got.Title)
	assert.Equal(t, want.Description, got.Description)
	assert.Equal(t, want.Location, got.Location)
	assert.True(t, want.Date.Equal(got.Date), "date %v, want %v", got.Date, want.Date)
	assert.Equal(t, want.AvailableQuota, got.AvailableQuota)
	assert.Equal(t, want.MaxQuota, got.MaxQuota)
	assert.Equal(t, want.Status, got.Status)
}

// assertInvalid accepts both codes a validation failure can come back as.
func assertInvalid(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()

	assert.Contains(t, []int{http.StatusBadRequest, http.StatusUnprocessableEntity}, w.Code, w.Body.String())
}
