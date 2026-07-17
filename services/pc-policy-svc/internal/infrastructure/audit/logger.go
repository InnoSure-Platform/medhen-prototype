package audit

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LogEntry struct {
	ID         uuid.UUID
	ActorID    string
	Action     string
	EntityType string
	EntityID   string
	Details    interface{}
}

type PostgresAuditLogger struct {
	pool *pgxpool.Pool
}

func NewPostgresAuditLogger(pool *pgxpool.Pool) *PostgresAuditLogger {
	return &PostgresAuditLogger{pool: pool}
}

// Log asynchronously records an audit event to comply with NBE regulations.
func (l *PostgresAuditLogger) Log(ctx context.Context, entry LogEntry) {
	// In a high-throughput environment, this could write to a local buffer or Kafka first.
	go func() {
		// Use a detached background context since the parent HTTP context might be cancelled
		bgCtx := context.Background()

		detailsJSON, _ := json.Marshal(entry.Details)

		query := `
			INSERT INTO audit_log (id, actor_id, action, entity_type, entity_id, details)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, _ = l.pool.Exec(bgCtx, query,
			uuid.New(), entry.ActorID, entry.Action, entry.EntityType, entry.EntityID, detailsJSON)
	}()
}
