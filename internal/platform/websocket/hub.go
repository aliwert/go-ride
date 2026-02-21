package websocket

import (
	"log"
	"sync"

	ws "github.com/gofiber/contrib/websocket"
)

// manages WebSocket connections grouped by trip ID.
// riders subscribe to a trip and receive real-time driver location pushes.
type Hub struct {
	mu    sync.RWMutex
	conns map[string][]*ws.Conn
}

func NewHub() *Hub {
	return &Hub{
		conns: make(map[string][]*ws.Conn),
	}
}

func (h *Hub) Register(tripID string, conn *ws.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.conns[tripID] = append(h.conns[tripID], conn)
}

func (h *Hub) Unregister(tripID string, conn *ws.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()

	conns := h.conns[tripID]
	for i, c := range conns {
		if c == conn {
			// swap-delete to avoid shifting the slice
			conns[i] = conns[len(conns)-1]
			h.conns[tripID] = conns[:len(conns)-1]
			break
		}
	}

	// clean up the map entry when no subscribers remain
	if len(h.conns[tripID]) == 0 {
		delete(h.conns, tripID)
	}
}

// sends a message to every connection watching a given trip.
// broken connections are silently removed so one bad client doesn't poison others.
func (h *Hub) BroadcastToTrip(tripID string, message []byte) {
	h.mu.RLock()
	conns := make([]*ws.Conn, len(h.conns[tripID]))
	copy(conns, h.conns[tripID])
	h.mu.RUnlock()

	var stale []*ws.Conn
	for _, c := range conns {
		if err := c.WriteMessage(ws.TextMessage, message); err != nil {
			log.Printf("WARN: ws write failed for trip %s: %v", tripID, err)
			stale = append(stale, c)
		}
	}

	// evict broken connections outside the read lock
	for _, c := range stale {
		h.Unregister(tripID, c)
	}
}
