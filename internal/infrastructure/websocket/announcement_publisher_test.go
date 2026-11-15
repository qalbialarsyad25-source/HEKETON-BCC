package websocket

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnnouncementPublisher_Publish(t *testing.T) {
	t.Run("broadcast message to clients subscribed to the same eventID", func(t *testing.T) {
		hub := NewHub()
		publisher := NewAnnouncementPublisher(hub)

		targetEventID := uuid.New()
		otherEventID := uuid.New()
		expectedPayload := []byte(`{"type":"announcement.created","data":{"title":"Info Penting"}}`)

		targetClient := &Client{
			EventID: targetEventID,
			Send:    make(chan []byte, 1),
		}
		otherClient := &Client{
			EventID: otherEventID,
			Send:    make(chan []byte, 1),
		}

		hub.Register(targetClient)
		hub.Register(otherClient)

		publisher.Publish(targetEventID, expectedPayload)

		// Verifikasi targetClient menerima pesan
		select {
		case received := <-targetClient.Send:
			assert.Equal(t, expectedPayload, received)
		case <-time.After(500 * time.Millisecond):
			require.Fail(t, "timeout: targetClient tidak menerima pesan broadcast")
		}

		// Verifikasi otherClient tidak menerima pesan
		select {
		case msg := <-otherClient.Send:
			require.Failf(t, "unexpected message", "otherClient seharusnya tidak menerima pesan: %s", string(msg))
		case <-time.After(50 * time.Millisecond):
			// Aman
		}
	})
}