//go:build e2e

package event_test

import (
	"net/http"
	"testing"

	"main/internal/shared/pagination"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListEvents(t *testing.T) {
	a := newAPI(t)
	organizer := a.newUserToken(t)
	viewer := a.newUserToken(t)

	t.Run("any signed-in user sees the events", func(t *testing.T) {
		firstReq := newCreateRequest()
		firstReq.Date = futureDate(365 * 5)
		secondReq := newCreateRequest()
		secondReq.Date = futureDate(365*5 + 1)
		first := createEvent(t, a, organizer, firstReq)
		second := createEvent(t, a, organizer, secondReq)

		w := a.do(t, http.MethodGet, eventsPath+"?limit=100", nil, viewer)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		list := decodeList(t, w)
		assert.Subset(t, eventIDs(list.Data), []string{first.ID, second.ID})
		assert.Equal(t, 1, list.Meta.Page)
		assert.Equal(t, 100, list.Meta.Limit)
		assert.GreaterOrEqual(t, list.Meta.TotalItems, int64(2))
	})

	t.Run("the page size is capped", func(t *testing.T) {
		w := a.do(t, http.MethodGet, eventsPath+"?limit=1000", nil, viewer)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		assert.Equal(t, pagination.MaxLimit, decodeList(t, w).Meta.Limit)
	})

	t.Run("invalid page and limit fall back to the defaults", func(t *testing.T) {
		w := a.do(t, http.MethodGet, eventsPath+"?page=abc&limit=-3", nil, viewer)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		list := decodeList(t, w)
		assert.Equal(t, pagination.DefaultPage, list.Meta.Page)
		assert.Equal(t, pagination.DefaultLimit, list.Meta.Limit)
		assert.LessOrEqual(t, len(list.Data), pagination.DefaultLimit)
	})

	t.Run("a page past the end is empty, not an error", func(t *testing.T) {
		w := a.do(t, http.MethodGet, eventsPath+"?page=999999", nil, viewer)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		list := decodeList(t, w)
		assert.NotNil(t, list.Data, "an empty page is [], not null")
		assert.Empty(t, list.Data)
	})

	t.Run("requires a signed-in user", func(t *testing.T) {
		w := a.do(t, http.MethodGet, eventsPath, nil, "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestListMyEvents(t *testing.T) {
	a := newAPI(t)

	t.Run("is empty for a user without events", func(t *testing.T) {
		token := a.newUserToken(t)

		w := a.do(t, http.MethodGet, myEventsPath, nil, token)

		require.Equal(t, http.StatusOK, w.Code, w.Body.String())
		list := decodeList(t, w)
		assert.NotNil(t, list.Data, "an empty list is [], not null")
		assert.Empty(t, list.Data)
		assert.Equal(t, int64(0), list.Meta.TotalItems)
		assert.Equal(t, 0, list.Meta.TotalPages)
	})

	t.Run("lists only the events the user belongs to", func(t *testing.T) {
		alice := a.newUserToken(t)
		bob := a.newUserToken(t)
		first := createEvent(t, a, alice, newCreateRequest())
		second := createEvent(t, a, alice, newCreateRequest())
		third := createEvent(t, a, bob, newCreateRequest())

		aliceList := decodeList(t, a.do(t, http.MethodGet, myEventsPath, nil, alice))
		bobList := decodeList(t, a.do(t, http.MethodGet, myEventsPath, nil, bob))

		assert.ElementsMatch(t, []string{first.ID, second.ID}, eventIDs(aliceList.Data))
		assert.Equal(t, int64(2), aliceList.Meta.TotalItems)
		assert.Equal(t, []string{third.ID}, eventIDs(bobList.Data))
	})

	t.Run("pages through the events, latest date first", func(t *testing.T) {
		token := a.newUserToken(t)
		var events [3]string
		for i, days := range []int{20, 30, 10} {
			req := newCreateRequest()
			req.Date = futureDate(days)
			events[i] = createEvent(t, a, token, req).ID
		}
		in20, in30, in10 := events[0], events[1], events[2]

		firstPage := a.do(t, http.MethodGet, myEventsPath+"?page=1&limit=2", nil, token)
		secondPage := a.do(t, http.MethodGet, myEventsPath+"?page=2&limit=2", nil, token)

		require.Equal(t, http.StatusOK, firstPage.Code, firstPage.Body.String())
		require.Equal(t, http.StatusOK, secondPage.Code, secondPage.Body.String())
		first := decodeList(t, firstPage)
		second := decodeList(t, secondPage)
		assert.Equal(t, []string{in30, in20}, eventIDs(first.Data))
		assert.Equal(t, []string{in10}, eventIDs(second.Data))
		assert.Equal(t, pagination.Meta{Page: 1, Limit: 2, TotalItems: 3, TotalPages: 2}, first.Meta)
		assert.Equal(t, pagination.Meta{Page: 2, Limit: 2, TotalItems: 3, TotalPages: 2}, second.Meta)
	})

	t.Run("requires a signed-in user", func(t *testing.T) {
		w := a.do(t, http.MethodGet, myEventsPath, nil, "")

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
