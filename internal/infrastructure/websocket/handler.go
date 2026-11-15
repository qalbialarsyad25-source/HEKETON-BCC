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

func (h *Handler) ServeWebSocket(c *gin.Context) {
	eventID, err := uuid.Parse(c.Query("event_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid event ID",
		})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := NewClient(h.hub, conn, eventID)

	h.hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}
