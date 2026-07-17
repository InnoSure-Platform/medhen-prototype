package rest

import (
	"encoding/base64"
	"log/slog"
	"net/http"
	"github.com/google/uuid"
	"pc-notification-svc/internal/application/command"
)

var transparentPixel = "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNkYAAAAAYAAjCB0C8AAAAASUVORK5CYII="

type TrackingServer struct {
	logger *slog.Logger
	receiptHandler *command.HandleDeliveryReceiptHandler
}

func NewTrackingServer(logger *slog.Logger, rh *command.HandleDeliveryReceiptHandler) *TrackingServer {
	return &TrackingServer{
		logger: logger,
		receiptHandler: rh,
	}
}

func (s *TrackingServer) TrackOpen(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	notifID, err := uuid.Parse(idStr)
	if err == nil {
		s.logger.Info("Email Opened", "notification_id", notifID)
		// Ideally publish to Kafka "NotificationOpened"
	}

	pixelData, _ := base64.StdEncoding.DecodeString(transparentPixel)
	w.Header().Set("Content-Type", "image/png")
	w.Write(pixelData)
}

func (s *TrackingServer) TrackClick(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	targetURL := r.URL.Query().Get("url")
	
	notifID, err := uuid.Parse(idStr)
	if err == nil {
		s.logger.Info("Email Link Clicked", "notification_id", notifID, "url", targetURL)
		// Ideally publish to Kafka "NotificationClicked"
	}

	if targetURL != "" {
		http.Redirect(w, r, targetURL, http.StatusFound)
		return
	}
	http.Error(w, "Bad Request", http.StatusBadRequest)
}
