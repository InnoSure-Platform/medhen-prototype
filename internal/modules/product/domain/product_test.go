package domain_test

import (
	"testing"

	"github.com/InnoSure-Platform/medhen-prototype/internal/modules/product/domain"
	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/money"
)

func cover(code string) domain.Coverage {
	return domain.Coverage{Code: code, Name: code + " cover", BaseRate: money.FromInt(1000)}
}

func TestNewProduct_Valid(t *testing.T) {
	p, err := domain.NewProduct("MOT", "MOTOR", "Motor", "የተሽከርካሪ", "v1",
		[]domain.Coverage{cover("OD")}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Status != domain.StatusActive || len(p.Coverages) != 1 {
		t.Fatalf("unexpected product: %+v", p)
	}
}

func TestNewProduct_Validation(t *testing.T) {
	if _, err := domain.NewProduct("", "MOTOR", "M", "", "v1", []domain.Coverage{cover("OD")}, nil); err != domain.ErrCodeRequired {
		t.Fatalf("empty code: got %v", err)
	}
	if _, err := domain.NewProduct("MOT", "MOTOR", "M", "", "v1", nil, nil); err != domain.ErrNoCoverages {
		t.Fatalf("no coverages: got %v", err)
	}
	bad := domain.Coverage{Code: "", Name: ""}
	if _, err := domain.NewProduct("MOT", "MOTOR", "M", "", "v1", []domain.Coverage{bad}, nil); err != domain.ErrCoverageInvalid {
		t.Fatalf("bad coverage: got %v", err)
	}
}
