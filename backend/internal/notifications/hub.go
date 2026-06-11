package notifications

import (
	"encoding/json"
	"fmt"
	"sync"

	"stock-market/backend/internal/repositories"
)

type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[chan []byte]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[chan []byte]struct{}),
	}
}

func (h *Hub) Subscribe(userID string) chan []byte {
	ch := make(chan []byte, 16)

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[userID] == nil {
		h.clients[userID] = make(map[chan []byte]struct{})
	}
	h.clients[userID][ch] = struct{}{}

	return ch
}

func (h *Hub) Unsubscribe(userID string, ch chan []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()

	subs, ok := h.clients[userID]
	if !ok {
		return
	}

	delete(subs, ch)
	close(ch)

	if len(subs) == 0 {
		delete(h.clients, userID)
	}
}

func (h *Hub) PublishNotification(userID string, notification repositories.Notification, unreadCount int) {
	payload, err := json.Marshal(map[string]any{
		"notification": notification,
		"unreadCount":  unreadCount,
	})
	if err != nil {
		return
	}

	h.publish(userID, formatSSE("notification", string(payload)))
}

func (h *Hub) publish(userID string, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	subs := h.clients[userID]
	for ch := range subs {
		select {
		case ch <- message:
		default:
		}
	}
}

func FormatSSE(event, data string) []byte {
	return []byte(fmt.Sprintf("event: %s\ndata: %s\n\n", event, data))
}

func formatSSE(event, data string) []byte {
	return FormatSSE(event, data)
}
