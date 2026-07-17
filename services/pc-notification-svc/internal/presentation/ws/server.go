package ws

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all for prototype
	},
}

type Hub struct {
	logger  *slog.Logger
	clients map[uuid.UUID]*websocket.Conn
	mu      sync.RWMutex
}

func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		logger:  logger,
		clients: make(map[uuid.UUID]*websocket.Conn),
	}
}

func (h *Hub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	partyIDStr := r.URL.Query().Get("party_id")
	partyID, err := uuid.Parse(partyIDStr)
	if err != nil {
		http.Error(w, "Invalid party_id", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade websocket", "error", err)
		return
	}

	h.mu.Lock()
	h.clients[partyID] = conn
	h.mu.Unlock()

	h.logger.Info("Client connected", "partyID", partyID)

	defer func() {
		h.mu.Lock()
		delete(h.clients, partyID)
		h.mu.Unlock()
		conn.Close()
		h.logger.Info("Client disconnected", "partyID", partyID)
	}()

	// Keep alive loop
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}

func (h *Hub) PushNotification(partyID uuid.UUID, payload interface{}) bool {
	h.mu.RLock()
	conn, ok := h.clients[partyID]
	h.mu.RUnlock()

	if !ok {
		return false // client offline
	}

	msg, _ := json.Marshal(payload)
	err := conn.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		h.logger.Error("Failed to push websocket message", "error", err)
		return false
	}
	return true
}
