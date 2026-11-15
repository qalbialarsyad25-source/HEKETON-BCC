package controller

import (
	"encoding/json"

	ws "main/internal/infrastructure/websocket"
	"main/internal/modules/session/domain"
	"main/internal/modules/session/dto"
	"main/internal/modules/session/usecase"

	"github.com/google/uuid"
)

// Realtime message types sent on GET /ws/events/:eventID/sessions.
const (
	MessageSessionCreated = "session.created"
	MessageSessionUpdated = "session.updated"
	MessageSessionDeleted = "session.deleted"
)

type message struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

// Broadcaster pushes session changes to the WebSocket clients that follow the session's event.
type Broadcaster struct {
	hub *ws.Hub
}

func NewBroadcaster(hub *ws.Hub) *Broadcaster {
	return &Broadcaster{hub: hub}
}

func (b *Broadcaster) SessionCreated(session domain.Session) {
	b.publish(session.EventID, MessageSessionCreated, session)
}

func (b *Broadcaster) SessionUpdated(session domain.Session) {
	b.publish(session.EventID, MessageSessionUpdated, session)
}

// SessionDeleted sends the session as it was before it was removed.
func (b *Broadcaster) SessionDeleted(session domain.Session) {
	b.publish(session.EventID, MessageSessionDeleted, session)
}

func (b *Broadcaster) publish(eventID uuid.UUID, messageType string, session domain.Session) {
	payload, err := json.Marshal(message{Type: messageType, Data: dto.SessionToResponse(session)})
	if err != nil {
		return
	}

	b.hub.Broadcast(eventID, payload)
}

var _ usecase.Notifier = (*Broadcaster)(nil)
