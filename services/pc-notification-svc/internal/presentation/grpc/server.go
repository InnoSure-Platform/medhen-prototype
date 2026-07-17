package grpc

import (
	"context"
	"log/slog"
	"net"

	"pc-notification-svc/internal/application/command"
	"github.com/google/uuid"
)

type Server struct {
	logger          *slog.Logger
	dispatchHandler *command.DispatchHandler
	receiptHandler  *command.HandleDeliveryReceiptHandler
}

func NewServer(logger *slog.Logger, dh *command.DispatchHandler, rh *command.HandleDeliveryReceiptHandler) *Server {
	return &Server{
		logger:          logger,
		dispatchHandler: dh,
		receiptHandler:  rh,
	}
}

// Mock gRPC method for DispatchAdHoc
func (s *Server) DispatchAdHoc(ctx context.Context, tenantID string, partyID uuid.UUID, tplCode string, payload map[string]interface{}) error {
	s.logger.Info("Received AdHoc dispatch", "partyID", partyID, "template", tplCode)
	
	cmd := command.DispatchCommand{
		TenantID:     tenantID,
		PartyID:      partyID,
		EventName:    tplCode, // Using template code as pseudo event name
		Payload:      payload,
		TargetLocale: "en-US",
	}

	return s.dispatchHandler.Handle(ctx, cmd)
}

func (s *Server) Start(port string) error {
	s.logger.Info("gRPC Server listening", "port", port)
	// mock start
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	defer listener.Close()
	return nil
}
