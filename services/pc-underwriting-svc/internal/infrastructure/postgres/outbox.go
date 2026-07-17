package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/medhen/pc-underwriting-svc/internal/application/port"
)

type OutboxRepo struct {
	db *pgxpool.Pool
}

func NewOutboxRepo(db *pgxpool.Pool) port.OutboxRepository {
	return &OutboxRepo{db: db}
}

func (r *OutboxRepo) PublishEvent(ctx context.Context, topic, partitionKey string, payload []byte) error {
	// Transactional outbox pattern
	query := `INSERT INTO outbox (topic, partition_key, payload, headers, created_at)
			  VALUES ($1, $2, $3, $4, $5)`
	
	headers, _ := json.Marshal(map[string]string{"source": "pc-underwriting-svc", "type": topic})
	
	_, err := r.db.Exec(ctx, query, topic, partitionKey, payload, headers, time.Now().UTC())
	return err
}
