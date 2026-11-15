package websocket

import "github.com/google/uuid"

type AnnouncementPublisher struct {
	hub *Hub
}

func NewAnnouncementPublisher(hub *Hub) *AnnouncementPublisher {
	return &AnnouncementPublisher{
		hub: hub,
	}
}

func (p *AnnouncementPublisher) Publish(eventID uuid.UUID, message []byte) {
	p.hub.Broadcast(eventID, message)
}