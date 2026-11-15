
package websocket

import (
	"sync"

	"github.com/google/uuid"
)

type Hub struct {
	mu      sync.RWMutex
	clients map[uuid.UUID]map[*Client]bool
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[uuid.UUID]map[*Client]bool),
	}
}

func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[client.EventID] == nil {
		h.clients[client.EventID] = make(map[*Client]bool)
	}

	h.clients[client.EventID][client] = true
}

func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	eventClients := h.clients[client.EventID]

	if eventClients == nil {
		return
	}

	if _, exists := eventClients[client]; exists {
		delete(eventClients, client)
		close(client.Send)
	}

	if len(eventClients) == 0 {
		delete(h.clients, client.EventID)
	}
}

func (h *Hub) Broadcast(eventID uuid.UUID, message []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	eventClients := h.clients[eventID]

	for client := range eventClients {
		select {
		case client.Send <- message:
		default:
			delete(eventClients, client)
			close(client.Send)
		}
	}

	if len(eventClients) == 0 {
		delete(h.clients, eventID)
	}
}
