package command

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"github.com/medhen/pc-underwriting-svc/internal/application/port"
	"github.com/medhen/pc-underwriting-svc/internal/domain/aggregate"
	"golang.org/x/sync/errgroup"
)

type AssessRiskCommand struct {
	TenantID    string
	QuoteID     uuid.UUID
	ProductID   string
	ProductLOB  string
	RiskPayload []byte
	Premium     float64
	TSI         float64
}

type AssessRiskHandler struct {
	uow            port.UnitOfWork
	assessmentRepo port.AssessmentRepository
	referralRepo   port.ReferralRepository
	outboxRepo     port.OutboxRepository
	productSvc     port.ProductServiceClient
	enrichment     port.EnrichmentProvider
}

func NewAssessRiskHandler(uow port.UnitOfWork, ar port.AssessmentRepository, rr port.ReferralRepository, or port.OutboxRepository, ps port.ProductServiceClient, ep port.EnrichmentProvider) *AssessRiskHandler {
	return &AssessRiskHandler{
		uow:            uow,
		assessmentRepo: ar,
		referralRepo:   rr,
		outboxRepo:     or,
		productSvc:     ps,
		enrichment:     ep,
	}
}

func (h *AssessRiskHandler) Handle(ctx context.Context, cmd AssessRiskCommand) (*aggregate.UnderwritingAssessment, error) {
	ctx, span := otel.Tracer("pc-underwriting-svc").Start(ctx, "AssessRiskHandler.Handle")
	defer span.End()

	// 1. Scatter-Gather Enrichment
	enrichCtx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()
	
	g, gCtx := errgroup.WithContext(enrichCtx)

	var priorClaims port.EnrichmentData
	var creditScore port.EnrichmentData

	if h.enrichment != nil {
		g.Go(func() error {
			var err error
			priorClaims, err = h.enrichment.FetchPriorClaims(gCtx, cmd.TenantID, cmd.QuoteID)
			return err
		})
		g.Go(func() error {
			var err error
			creditScore, err = h.enrichment.FetchCreditScore(gCtx, cmd.TenantID, cmd.QuoteID)
			return err
		})
	}

	if err := g.Wait(); err != nil {
		// Log warning, but we can potentially proceed without enrichment or fail
		// log.Printf("Enrichment failed: %v", err)
	}

	// Mock merging enrichment data into payload
	// cmd.RiskPayload = merge(cmd.RiskPayload, priorClaims, creditScore)
	_ = priorClaims
	_ = creditScore

	// 2. Evaluate DMN Rules via Product Service
	status, score, rules, err := h.productSvc.EvaluateDMNRules(ctx, cmd.ProductID, cmd.RiskPayload)
	if err != nil {
		return nil, err
	}

	// 2. Create Assessment
	assessment, err := aggregate.NewUnderwritingAssessment(cmd.TenantID, cmd.QuoteID, cmd.ProductID, score, rules, status)
	if err != nil {
		return nil, err
	}

	// 3. Persist transactionally
	err = h.uow.Do(ctx, func(txCtx context.Context) error {
		if err := h.assessmentRepo.Save(txCtx, assessment); err != nil {
			return err
		}

		if status == aggregate.StatusReferred {
			// Facultative check (mock logic: TSI > 50M requires facultative)
			isFacultative := cmd.TSI > 50000000.0

			// Create Referral
			referral := aggregate.NewReferral(cmd.TenantID, assessment.ID, "L1", 24, isFacultative)
			if err := h.referralRepo.Save(txCtx, referral); err != nil {
				return err
			}

			// Publish ReferralCreated Event
			payload, _ := json.Marshal(map[string]interface{}{"referral_id": referral.ID, "assessment_id": assessment.ID})
			if err := h.outboxRepo.PublishEvent(txCtx, "pc.underwriting.referral.lifecycle.v1", referral.ID.String(), payload); err != nil {
				return err
			}
		}

		// Publish RiskAssessed Event
		eventPayload, _ := json.Marshal(map[string]interface{}{"quote_id": assessment.QuoteID, "status": assessment.Status, "score": assessment.RiskScore.Value})
		if err := h.outboxRepo.PublishEvent(txCtx, "pc.underwriting.risk.assessed.v1", assessment.QuoteID.String(), eventPayload); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return assessment, nil
}
