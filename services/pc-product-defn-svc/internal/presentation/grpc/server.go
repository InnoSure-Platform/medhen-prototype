package grpc

import (
	"context"
	"errors"

	productpb "github.com/medhen/pc-contracts/gen/go/product/v1"
	"github.com/medhen/pc-product-defn-svc/internal/application/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	productpb.UnimplementedProductQueryServiceServer
	getProductQuery *query.GetProductHandler
}

func NewServer(getProductQuery *query.GetProductHandler) *Server {
	return &Server{
		getProductQuery: getProductQuery,
	}
}

func (s *Server) GetEffectiveProduct(ctx context.Context, req *productpb.GetEffectiveProductRequest) (*productpb.GetEffectiveProductResponse, error) {
	if req.ProductId == "" {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}

	// For the sake of this Tier-0 hardening, we'll fetch the product using the existing query handler.
	// In a fully featured system, this might use the target_timestamp to fetch an older version.
	product, err := s.getProductQuery.Handle(ctx, req.ProductId)
	if err != nil {
		if errors.Is(err, query.ErrProductNotFound) {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Map coverages
	var pbCoverages []*productpb.Coverage
	for _, cov := range product.Coverages {
		pbCoverages = append(pbCoverages, &productpb.Coverage{
			Code:        cov.Code,
			Name:        cov.Name,
			IsMandatory: cov.IsMandatory,
		})
	}

	return &productpb.GetEffectiveProductResponse{
		Id:        product.ID.String(),
		Code:      product.Code,
		Name:      product.Name,
		Version:   int32(product.Version),
		Coverages: pbCoverages,
	}, nil
}
