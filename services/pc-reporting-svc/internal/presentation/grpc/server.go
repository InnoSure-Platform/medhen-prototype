package grpc

import (
	"context"
	
	"github.com/medhen/pc-contracts/gen/go/reporting/v1"
	"github.com/medhen/pc-reporting-svc/internal/application/query"
	"go.opentelemetry.io/otel"
)

var tracer = otel.Tracer("medhen.com/pc-reporting-svc/grpc")

type Server struct {
	reportingpb.UnimplementedReportingQueryServiceServer
	kpiQuery *query.KPIHandler
}

func NewServer(kpiQuery *query.KPIHandler) *Server {
	return &Server{kpiQuery: kpiQuery}
}

func (s *Server) GetProductionKPIs(ctx context.Context, req *reportingpb.GetProductionKPIsRequest) (*reportingpb.GetProductionKPIsResponse, error) {
	ctx, span := tracer.Start(ctx, "GetProductionKPIs_gRPC")
	defer span.End()

	summary, err := s.kpiQuery.Handle(ctx, req.TenantId, req.Lob)
	if err != nil {
		return nil, err
	}

	return &reportingpb.GetProductionKPIsResponse{
		TotalGwp:      summary.TotalGWP,
		TotalNwp:      summary.TotalNWP,
		PolicyCount:   summary.PolicyCount,
		LossRatio:     summary.LossRatio,
		CombinedRatio: summary.CombinedRatio,
	}, nil
}
