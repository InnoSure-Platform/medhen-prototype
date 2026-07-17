package grpc

import (
	"context"

	"github.com/google/uuid"
	pb "github.com/medhen/pc-contracts/gen/go/integration/v1"
	"github.com/medhen/pc-integration-svc/internal/application/command"
	"github.com/medhen/pc-integration-svc/internal/infrastructure/providers"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements the IntegrationService gRPC interface
type Server struct {
	pb.UnimplementedIntegrationServiceServer
	logger          *zap.Logger
	initiatePayment *command.InitiatePaymentHandler
	// Using the adapters directly for now for simplicity in MVP, but they should be wrapped in use cases
	faydaClient *providers.FaydaClient
	smsClient   *providers.SMSClient
}

func NewServer(
	logger *zap.Logger,
	initiatePayment *command.InitiatePaymentHandler,
	faydaClient *providers.FaydaClient,
	smsClient *providers.SMSClient,
) *Server {
	return &Server{
		logger:          logger,
		initiatePayment: initiatePayment,
		faydaClient:     faydaClient,
		smsClient:       smsClient,
	}
}

func (s *Server) InitiatePayment(ctx context.Context, req *pb.InitiatePaymentRequest) (*pb.InitiatePaymentResponse, error) {
	internalRef, err := uuid.Parse(req.InternalReferenceId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid internal reference id: %v", err)
	}

	cmd := command.InitiatePaymentCmd{
		InternalReferenceID: internalRef,
		Provider:            req.Provider,
		Amount:              req.Amount,
		Currency:            req.Currency,
	}

	redirectURL, err := s.initiatePayment.Handle(ctx, cmd)
	if err != nil {
		s.logger.Error("failed to initiate payment", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to initiate payment: %v", err)
	}

	return &pb.InitiatePaymentResponse{
		RedirectUrl: redirectURL,
	}, nil
}

func (s *Server) VerifyIdentity(ctx context.Context, req *pb.VerifyIdentityRequest) (*pb.VerifyIdentityResponse, error) {
	isVerified, reason, err := s.faydaClient.VerifyIdentity(ctx, req.FaydaId, req.FirstName, req.LastName, req.DateOfBirth)
	if err != nil {
		s.logger.Error("failed to verify identity", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to verify identity: %v", err)
	}

	return &pb.VerifyIdentityResponse{
		IsVerified:            isVerified,
		VerificationReference: req.FaydaId,
		MismatchReason:        reason,
	}, nil
}

func (s *Server) SendSMS(ctx context.Context, req *pb.SendSMSRequest) (*pb.SendSMSResponse, error) {
	success, msgID, err := s.smsClient.SendSMS(ctx, req.PhoneNumber, req.Message)
	if err != nil {
		s.logger.Error("failed to send sms", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to send sms: %v", err)
	}

	return &pb.SendSMSResponse{
		Success:   success,
		MessageId: msgID,
	}, nil
}
