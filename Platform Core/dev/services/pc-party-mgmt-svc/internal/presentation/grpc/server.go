package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	partypb "github.com/medhen/pc-contracts/gen/go/party/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PartyServer implements the generated gRPC PartyResolutionServiceServer interface.
type PartyServer struct {
	partypb.UnimplementedPartyResolutionServiceServer
	db *pgxpool.Pool
}

// NewPartyServer creates a new gRPC server for Party Management.
func NewPartyServer(db *pgxpool.Pool) *PartyServer {
	return &PartyServer{
		db: db,
	}
}

// ResolveParty handles internal gRPC requests for Party Summaries.
func (s *PartyServer) ResolveParty(ctx context.Context, req *partypb.ResolvePartyRequest) (*partypb.PartySummary, error) {
	partyID, err := uuid.Parse(req.PartyId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid party id")
	}

	query := `
		SELECT type, first_name, last_name, legal_name, status 
		FROM parties 
		WHERE id = $1 AND tenant_id = $2
	`
	
	var pType, firstName, lastName, legalName, pStatus *string
	err = s.db.QueryRow(ctx, query, partyID, req.TenantId).Scan(&pType, &firstName, &lastName, &legalName, &pStatus)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "party not found")
	}

	res := &partypb.PartySummary{
		Id: req.PartyId,
	}
	if pType != nil { res.Type = *pType }
	if firstName != nil { res.FirstName = *firstName }
	if lastName != nil { res.LastName = *lastName }
	if legalName != nil { res.LegalName = *legalName }
	if pStatus != nil { res.Status = *pStatus }

	return res, nil
}

func (s *PartyServer) VerifyKYCState(ctx context.Context, req *partypb.VerifyKYCStateRequest) (*partypb.KYCStatusResponse, error) {
	partyID, err := uuid.Parse(req.PartyId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid party id")
	}

	var kycStatus string
	query := `SELECT kyc_status FROM parties WHERE id = $1 AND tenant_id = $2`
	err = s.db.QueryRow(ctx, query, partyID, req.TenantId).Scan(&kycStatus)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "party not found")
	}

	return &partypb.KYCStatusResponse{
		Status: kycStatus,
	}, nil
}
