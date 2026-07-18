package extractor

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type NBEMotorExtractor struct {
	conn   driver.Conn
	sealer *CryptoSealer
}

func NewNBEMotorExtractor(conn driver.Conn, sealer *CryptoSealer) *NBEMotorExtractor {
	return &NBEMotorExtractor{
		conn:   conn,
		sealer: sealer,
	}
}

// Extract generates the Motor Third-Party regulatory return for Reg 554/2024
func (e *NBEMotorExtractor) Extract(ctx context.Context, tenantID string, year int, quarter int) (*SealedDocument, error) {
	// Dummy query structure for the report extraction
	query := `
		SELECT policy_number, effective_date, gross_written_premium 
		FROM rm_policies_fact 
		WHERE tenant_id = $1 AND lob = 'MOTOR_TP' AND toYear(effective_date) = $2
	`
	
	rows, err := e.conn.Query(ctx, query, tenantID, year)
	if err != nil {
		return nil, fmt.Errorf("failed to query ClickHouse for NBE extract: %w", err)
	}
	defer rows.Close()

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	
	// Write Header
	_ = writer.Write([]string{"Policy Number", "Effective Date", "GWP"})

	for rows.Next() {
		var (
			policyNumber  string
			effectiveDate string
			gwp           float64
		)
		if err := rows.Scan(&policyNumber, &effectiveDate, &gwp); err != nil {
			return nil, err
		}
		_ = writer.Write([]string{policyNumber, effectiveDate, fmt.Sprintf("%.2f", gwp)})
	}
	writer.Flush()

	payload := buf.Bytes()
	hash, err := e.sealer.Seal(payload)
	if err != nil {
		return nil, err
	}

	return &SealedDocument{
		Payload: payload,
		Hash:    hash,
	}, nil
}
