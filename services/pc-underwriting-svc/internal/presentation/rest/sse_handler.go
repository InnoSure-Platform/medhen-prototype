package rest

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// SSEHandler manages Server-Sent Events connections to push real-time updates to the workbench UI.
type SSEHandler struct {
	// In a real app, this would contain a map of connected clients
	// or a pub/sub mechanism (like Redis PubSub) to fan-out events.
}

func NewSSEHandler() *SSEHandler {
	return &SSEHandler{}
}

func (h *SSEHandler) HandleEvents(w http.ResponseWriter, r *http.Request) {
	// Ensure the connection supports flushing
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ctx := r.Context()

	// Push an initial connection established message
	fmt.Fprintf(w, "event: connect\ndata: {\"status\":\"connected\"}\n\n")
	flusher.Flush()

	// Simulate pushing real-time referral updates every 10 seconds
	// In reality, this loop would block on a Go channel waiting for Redis Pub/Sub events 
	// emitted by the Outbox processor when a new referral is created or SLA breached.
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return // Client disconnected
		case t := <-ticker.C:
			// Mocking a live ping/update
			payload := fmt.Sprintf(`{"time": "%s", "message": "ping"}`, t.Format(time.RFC3339))
			fmt.Fprintf(w, "event: update\ndata: %s\n\n", payload)
			flusher.Flush()
		}
	}
}

func (h *SSEHandler) RegisterRoutes(r chi.Router) {
	r.Get("/api/pc-underwriting/v1/events", h.HandleEvents)
}
