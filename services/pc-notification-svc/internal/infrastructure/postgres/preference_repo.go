package postgres

import (
	"context"
	"github.com/google/uuid"
	"pc-notification-svc/internal/domain/preference"
	// "github.com/jackc/pgx/v5/pgxpool"
)

type PreferenceRepo struct {
	// pool *pgxpool.Pool
}

func NewPreferenceRepo() *PreferenceRepo {
	return &PreferenceRepo{}
}

func (r *PreferenceRepo) GetByPartyID(ctx context.Context, partyID uuid.UUID) (*preference.RoutingPreference, error) {
	/*
	query := `SELECT party_id, opted_out_sms, opted_out_email, opted_out_in_app, marketing_opt_in
			  FROM routing_preferences WHERE party_id = $1`
			  
	var p preference.RoutingPreference
	err := r.pool.QueryRow(ctx, query, partyID).Scan(
		&p.PartyID, &p.OptedOutSMS, &p.OptedOutEmail, &p.OptedOutInApp, &p.MarketingOptIn,
	)
	if err != nil {
		// If not found, default to false (opted-in to transactional)
		return &preference.RoutingPreference{PartyID: partyID}, nil
	}
	return &p, nil
	*/
	
	return &preference.RoutingPreference{
		PartyID: partyID,
	}, nil
}
