package controller

import (
	"encoding/json"
	"testing"
	"time"

	ws "main/internal/infrastructure/websocket"
	"main/internal/modules/session/factory"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBroadcaster(t *testing.T) {
	session := factory.NewSession(uuid.New())

	hub := ws.NewHub()
	follower := &ws.Client{EventID: session.EventID, Send: make(chan []byte, 3)}
	otherEvent := &ws.Client{EventID: uuid.New(), Send: make(chan []byte, 1)}
	hub.Register(follower)
	hub.Register(otherEvent)

	broadcaster := NewBroadcaster(hub)
	broadcaster.SessionCreated(*session)
	broadcaster.SessionUpdated(*session)
	broadcaster.SessionDeleted(*session)

	want := []string{MessageSessionCreated, MessageSessionUpdated, MessageSessionDeleted}
	for _, messageType := range want {
		select {
		case raw := <-follower.Send:
			var got struct {
				Type string `json:"type"`
				Data struct {
					ID      string `json:"id"`
					EventID string `json:"event_id"`
					Title   string `json:"title"`
					Status  string `json:"status"`
				} `json:"data"`
			}
			require.NoError(t, json.Unmarshal(raw, &got), string(raw))
			assert.Equal(t, messageType, got.Type)
			assert.Equal(t, session.ID.String(), got.Data.ID)
			assert.Equal(t, session.EventID.String(), got.Data.EventID)
			assert.Equal(t, session.Title, got.Data.Title)
			assert.Equal(t, string(session.Status), got.Data.Status)
		case <-time.After(500 * time.Millisecond):
			require.Failf(t, "timeout", "the follower did not get %s", messageType)
		}
	}

	select {
	case raw := <-otherEvent.Send:
		require.Failf(t, "unexpected message", "another event's client got: %s", string(raw))
	case <-time.After(50 * time.Millisecond):
	}
}
