package port

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/medhen/pc-underwriting-svc/internal/domain/aggregate"
)

// UnitOfWork defines a transactional boundary.
type UnitOfWork interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

type AssessmentRepository interface {
	Save(ctx context.Context, assessment *aggregate.UnderwritingAssessment) error
	FindByID(ctx context.Context, id uuid.UUID) (*aggregate.UnderwritingAssessment, error)
	FindByQuoteID(ctx context.Context, tenantID string, quoteID uuid.UUID) (*aggregate.UnderwritingAssessment, error)
}

type ReferralRepository interface {
	Save(ctx context.Context, referral *aggregate.Referral) error
	FindByID(ctx context.Context, id uuid.UUID) (*aggregate.Referral, error)
	Update(ctx context.Context, referral *aggregate.Referral) error
	FindBreachedSLAs(ctx context.Context, asOf time.Time, limit int) ([]*aggregate.Referral, error)
}

type AuthorityRepository interface {
	FindByLevelAndProduct(ctx context.Context, tenantID, levelCode, productLOB string) (*aggregate.AuthorityLevel, error)
}

type OutboxRepository interface {
	PublishEvent(ctx context.Context, topic, partitionKey string, payload []byte) error
}

type ProductServiceClient interface {
	EvaluateDMNRules(ctx context.Context, productID string, riskPayload []byte) (decision aggregate.AssessmentStatus, score int, rulesTriggered []string, err error)
}
