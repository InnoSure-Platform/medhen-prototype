package models_test

import (
	"testing"

	"medhen.com/platform/pc-rating-calc-svc/internal/domain/models"
)

func TestPremiumBreakdown_AddTrace(t *testing.T) {
	pb := &models.PremiumBreakdown{}

	pb.AddTrace("BASE_RATE", "100", "v1", "trace-1", "span-1")
	pb.AddTrace("AGE_FACTOR", "1.5", "v1", "trace-1", "span-2")

	if len(pb.TraceLog) != 2 {
		t.Fatalf("expected 2 traces, got %d", len(pb.TraceLog))
	}

	if pb.TraceLog[0].StepOrder != 1 || pb.TraceLog[0].Operation != "BASE_RATE" {
		t.Errorf("unexpected first trace: %+v", pb.TraceLog[0])
	}

	if pb.TraceLog[1].StepOrder != 2 || pb.TraceLog[1].Operation != "AGE_FACTOR" {
		t.Errorf("unexpected second trace: %+v", pb.TraceLog[1])
	}
}
