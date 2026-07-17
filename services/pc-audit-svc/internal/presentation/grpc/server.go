package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/medhen/pc-audit-svc/internal/application/commands"
	"github.com/medhen/pc-audit-svc/internal/domain/audit"
)

// AuditIngestionServer implements the medhen.platform.audit.v1.AuditIngestionService
type AuditIngestionServer struct {
	// UnimplementedAuditIngestionServiceServer
	appendHandler *commands.AppendRecordHandler
}

func NewAuditIngestionServer(appendHandler *commands.AppendRecordHandler) *AuditIngestionServer {
	return &AuditIngestionServer{appendHandler: appendHandler}
}

func (s *AuditIngestionServer) Register(server *grpc.Server) {
	// pb.RegisterAuditIngestionServiceServer(server, s)
}

// LogAction is the synchronous ingestion endpoint for critical events (e.g. IAM, Auth)
func (s *AuditIngestionServer) LogAction(ctx context.Context, req interface{}) (interface{}, error) {
	// Parse request (stubbed due to missing pb definitions)

	cmd := commands.AppendRecordCommand{
		TenantID:       "t-123",
		Actor:          audit.ActorContext{UserID: "usr-1", Role: "admin", IPAddress: "127.0.0.1"},
		ActionType:     "LOGIN",
		EntityType:     "UserSession",
		EntityID:       "sess-1",
		IsPIIEncrypted: false,
		DeltaPlaintext: []byte(`{"status": "SUCCESS"}`),
	}

	if err := s.appendHandler.Handle(ctx, cmd); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to append audit record: %v", err)
	}

	return nil, nil // Return Response ACK
}
