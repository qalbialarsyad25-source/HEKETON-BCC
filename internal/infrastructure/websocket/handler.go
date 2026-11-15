package websocket

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Handler struct {
	hub        *Hub
	membership EventAccessChecker
}

type EventAccessChecker interface {
	IsInternalMember(
		ctx context.Context,
		userID uuid.UUID,
		eventID uuid.UUID,
	) (bool, error)
}

func NewHandler(
	hub *Hub,
	membership EventAccessChecker,
) *Handler {
	return &Handler{
		hub:        hub,
		membership: membership,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Hanya untuk development lokal.
	},
}

// ServeWebSocket reads the event from ?event_id= and subscribes the connection to it.
func (h *Handler) ServeWebSocket(c *gin.Context) {
	eventID, err := uuid.Parse(c.Query("event_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid event ID",
		})
		return
	}

	h.Serve(c, eventID)
}

// Serve upgrades the request to a WebSocket and subscribes the connection to eventID.
// The caller has already checked that the event exists and that the caller may follow it.
func (h *Handler) Serve(c *gin.Context, eventID uuid.UUID) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := NewClient(h.hub, conn, eventID)

	h.hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}
