package workflow

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/workflow"
)

type BindPolicyInput struct {
	QuoteID uuid.UUID
}

type BindPolicyResult struct {
	PolicyID uuid.UUID
}

func BindPolicyWorkflow(ctx workflow.Context, input BindPolicyInput) (*BindPolicyResult, error) {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Starting BindPolicyWorkflow", "QuoteID", input.QuoteID)

	var a *Activities

	// Step 1: Reserve Funds via pc-billing-svc
	err := workflow.ExecuteActivity(ctx, a.ReserveFundsActivity, input.QuoteID).Get(ctx, nil)
	if err != nil {
		logger.Error("ReserveFundsActivity failed", "Error", err)
		return nil, err
	}

	// Step 2: Finalize Bind in pc-policy-svc local DB
	var policyID uuid.UUID
	err = workflow.ExecuteActivity(ctx, a.FinalizeBindActivity, input.QuoteID).Get(ctx, &policyID)
	if err != nil {
		logger.Error("FinalizeBindActivity failed, executing compensation", "Error", err)
		
		// Compensation: Release funds if finalize fails
		compCtx, _ := workflow.NewDisconnectedContext(ctx)
		_ = workflow.ExecuteActivity(compCtx, a.ReleaseFundsActivity, input.QuoteID).Get(compCtx, nil)
		
		return nil, err
	}

	// Step 3: Trigger Document Generation
	err = workflow.ExecuteActivity(ctx, a.GenerateDocumentsActivity, policyID).Get(ctx, nil)
	if err != nil {
		logger.Warn("GenerateDocumentsActivity failed, but policy is bound", "Error", err)
		// We don't fail the saga here, documents can be generated later by a retry/cron
	}

	logger.Info("BindPolicyWorkflow completed successfully", "PolicyID", policyID)
	return &BindPolicyResult{PolicyID: policyID}, nil
}

type Activities struct {
	// In a real app, these would have dependencies injected (e.g. gRPC clients, repos)
}

func (a *Activities) ReserveFundsActivity(ctx context.Context, quoteID uuid.UUID) error {
	// Call pc-billing-svc to reserve funds
	return nil
}

func (a *Activities) FinalizeBindActivity(ctx context.Context, quoteID uuid.UUID) (uuid.UUID, error) {
	// Call the local policy DB to mark as BOUND and emit the outbox event
	return uuid.New(), nil
}

func (a *Activities) ReleaseFundsActivity(ctx context.Context, quoteID uuid.UUID) error {
	// Compensating action: Release reserved funds in pc-billing-svc
	return nil
}

func (a *Activities) GenerateDocumentsActivity(ctx context.Context, policyID uuid.UUID) error {
	// Mock: Call pc-document-mgmt-svc to generate PDF Schedule and CoI, then trigger notification
	return nil
}
