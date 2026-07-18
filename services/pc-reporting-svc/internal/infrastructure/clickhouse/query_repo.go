package clickhouse

import (
	"context"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/medhen/pc-auth-sdk"
	"github.com/medhen/pc-reporting-svc/internal/domain"
)

type QueryRepository struct {
	conn driver.Conn
}

func NewQueryRepository(conn driver.Conn) *QueryRepository {
	return &QueryRepository{conn: conn}
}

// GetKPISummary executes a sub-10ms query against the OLAP ClickHouse backend
// Integrates BC-MDH-16: Row-Level Security based on JWT claims
func (r *QueryRepository) GetKPISummary(ctx context.Context, tenantID string, lob string) (*domain.KPISummary, error) {
	claims, ok := auth.GetClaims(ctx)
	
	query := `
		SELECT 
			sum(gross_written_premium * sign) as total_gwp,
			sum(net_written_premium * sign) as total_nwp,
			count() as policy_count
		FROM rm_policies_fact
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}

	// Dynamic Data Masking / RLS: If the user belongs to a specific branch, restrict their view.
	if ok && claims.BranchCode != "" && claims.BranchCode != "ALL" {
		query += " AND branch_code = $2"
		args = append(args, claims.BranchCode)
	}

	// LOB Filtering
	if lob != "" {
		paramIdx := fmt.Sprintf("$%d", len(args)+1)
		query += " AND lob = " + paramIdx
		args = append(args, lob)
	}

	var summary domain.KPISummary
	row := r.conn.QueryRow(ctx, query, args...)
	if err := row.Scan(&summary.TotalGWP, &summary.TotalNWP, &summary.PolicyCount); err != nil {
		return nil, fmt.Errorf("failed to scan KPI summary: %w", err)
	}

	// Placeholder for Claims integration to calculate Loss Ratio
	summary.LossRatio = 0.65 // Dummy value for now
	summary.CombinedRatio = 0.95

	return &summary, nil
}
