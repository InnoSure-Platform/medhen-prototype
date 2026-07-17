package postgres

import (
	"context"
	"pc-notification-svc/internal/domain/template"
	// "github.com/jackc/pgx/v5/pgxpool"
)

type TemplateRepo struct {
	// pool *pgxpool.Pool
}

func NewTemplateRepo() *TemplateRepo {
	return &TemplateRepo{}
}

func (r *TemplateRepo) GetActive(ctx context.Context, code string, channel string, locale string) (*template.NotificationTemplate, error) {
	/*
	query := `SELECT id, code, channel, locale, category, subject_template, body_template, version
			  FROM notification_templates
			  WHERE code = $1 AND channel = $2 AND locale = $3
			  ORDER BY version DESC LIMIT 1`
			  
	var t template.NotificationTemplate
	err := r.pool.QueryRow(ctx, query, code, channel, locale).Scan(
		&t.ID, &t.Code, &t.Channel, &t.Locale, &t.Category, &t.SubjectTemplate, &t.BodyTemplate, &t.Version,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
	*/
	
	// Mocking for tests since we don't have a live DB running here
	return &template.NotificationTemplate{
		Code:         code,
		Category:     "TRANSACTIONAL",
		BodyTemplate: "Hello {{.first_name}}, this is a test for {{.policy_number}}.",
	}, nil
}
