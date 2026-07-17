package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/medhen/pc-party-mgmt-svc/internal/domain"
)

type DB interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type PartyRepository struct {
	db DB
}

func NewPartyRepository(db *pgxpool.Pool) *PartyRepository {
	return &PartyRepository{db: db}
}

// PartyTxRepository wraps the same logic but enforces it runs inside a tx
type PartyTxRepository struct {
	tx pgx.Tx
}

// Save executes an UPSERT
func saveParty(ctx context.Context, db DB, party *domain.Party) error {
	query := `
		INSERT INTO parties (
			id, tenant_id, type, status, kyc_status, first_name, last_name, 
			dob, gender, national_id_type, national_id_number, legal_name, 
			registration_number, industry_code, tin, surviving_party_id, version, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, 
			$8, $9, $10, $11, $12, 
			$13, $14, $15, $16, $17, $18, $19
		)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			kyc_status = EXCLUDED.kyc_status,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			dob = EXCLUDED.dob,
			gender = EXCLUDED.gender,
			national_id_type = EXCLUDED.national_id_type,
			national_id_number = EXCLUDED.national_id_number,
			legal_name = EXCLUDED.legal_name,
			registration_number = EXCLUDED.registration_number,
			industry_code = EXCLUDED.industry_code,
			tin = EXCLUDED.tin,
			surviving_party_id = EXCLUDED.surviving_party_id,
			version = parties.version + 1,
			updated_at = EXCLUDED.updated_at
		WHERE parties.version = EXCLUDED.version - 1;
	`

	tag, err := db.Exec(ctx, query,
		party.ID, party.TenantID, party.Type, party.Status, party.KYCStatus,
		party.FirstName, party.LastName, party.DOB, party.Gender, party.NationalIDType, party.NationalIDNumber,
		party.LegalName, party.RegistrationNo, party.IndustryCode, party.TIN, party.SurvivingPartyID,
		party.Version, party.CreatedAt, party.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save party: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return errors.New("optimistic locking failure or duplicate issue")
	}

	for _, c := range party.Consents {
		cQuery := `
			INSERT INTO party_consents (party_id, consent_type, status, version, updated_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (party_id, consent_type) DO UPDATE SET
				status = EXCLUDED.status,
				version = EXCLUDED.version,
				updated_at = EXCLUDED.updated_at
			WHERE party_consents.version < EXCLUDED.version;
		`
		_, err = db.Exec(ctx, cQuery, c.PartyID, c.ConsentType, c.Status, c.Version, c.UpdatedAt)
		if err != nil {
			return fmt.Errorf("failed to save consent: %w", err)
		}
	}

	return nil
}

func (r *PartyRepository) Save(ctx context.Context, party *domain.Party) error {
	return saveParty(ctx, r.db, party)
}

func (r *PartyTxRepository) Save(ctx context.Context, party *domain.Party) error {
	return saveParty(ctx, r.tx, party)
}

func (r *PartyRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Party, error) {
	return findByID(ctx, r.db, id)
}

func (r *PartyTxRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Party, error) {
	return findByID(ctx, r.tx, id)
}

func findByID(ctx context.Context, db DB, id uuid.UUID) (*domain.Party, error) {
	query := `
		SELECT 
			id, tenant_id, type, status, kyc_status, first_name, last_name, 
			dob, gender, national_id_type, national_id_number, legal_name, 
			registration_number, industry_code, tin, surviving_party_id, version, created_at, updated_at
		FROM parties 
		WHERE id = $1
	`
	row := db.QueryRow(ctx, query, id)
	p, err := scanParty(row)
	if err != nil {
		return nil, err
	}
	if err := loadConsents(ctx, db, p); err != nil {
		return nil, fmt.Errorf("failed to load consents: %w", err)
	}
	return p, nil
}

func (r *PartyRepository) FindByNationalID(ctx context.Context, tenantID, nationalID string) (*domain.Party, error) {
	return findByNationalID(ctx, r.db, tenantID, nationalID)
}

func (r *PartyTxRepository) FindByNationalID(ctx context.Context, tenantID, nationalID string) (*domain.Party, error) {
	return findByNationalID(ctx, r.tx, tenantID, nationalID)
}

func findByNationalID(ctx context.Context, db DB, tenantID, nationalID string) (*domain.Party, error) {
	query := `
		SELECT 
			id, tenant_id, type, status, kyc_status, first_name, last_name, 
			dob, gender, national_id_type, national_id_number, legal_name, 
			registration_number, industry_code, tin, surviving_party_id, version, created_at, updated_at
		FROM parties 
		WHERE tenant_id = $1 AND national_id_number = $2 AND status != 'MERGED'
		LIMIT 1
	`
	row := db.QueryRow(ctx, query, tenantID, nationalID)
	p, err := scanParty(row)
	if err != nil {
		return nil, err
	}
	if err := loadConsents(ctx, db, p); err != nil {
		return nil, fmt.Errorf("failed to load consents: %w", err)
	}
	return p, nil
}

func scanParty(row pgx.Row) (*domain.Party, error) {
	var p domain.Party
	var dob *time.Time
	var survivingPartyID *uuid.UUID
	var firstName, lastName, gender, natIDType, natIDNum, legalName, regNo, indCode, tin *string

	err := row.Scan(
		&p.ID, &p.TenantID, &p.Type, &p.Status, &p.KYCStatus,
		&firstName, &lastName, &dob, &gender, &natIDType, &natIDNum,
		&legalName, &regNo, &indCode, &tin,
		&survivingPartyID, &p.Version, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("party not found")
		}
		return nil, fmt.Errorf("error scanning party: %w", err)
	}

	if firstName != nil { p.FirstName = *firstName }
	if lastName != nil { p.LastName = *lastName }
	p.DOB = dob
	if gender != nil { p.Gender = *gender }
	if natIDType != nil { p.NationalIDType = *natIDType }
	if natIDNum != nil { p.NationalIDNumber = *natIDNum }
	if legalName != nil { p.LegalName = *legalName }
	if regNo != nil { p.RegistrationNo = *regNo }
	if indCode != nil { p.IndustryCode = *indCode }
	if tin != nil { p.TIN = *tin }
	p.SurvivingPartyID = survivingPartyID

	return &p, nil
}

func loadConsents(ctx context.Context, db DB, p *domain.Party) error {
	query := `SELECT consent_type, status, version, updated_at FROM party_consents WHERE party_id = $1`
	rows, err := db.Query(ctx, query, p.ID)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var c domain.ConsentRecord
		c.PartyID = p.ID
		err := rows.Scan(&c.ConsentType, &c.Status, &c.Version, &c.UpdatedAt)
		if err != nil {
			return err
		}
		p.Consents = append(p.Consents, c)
	}
	return rows.Err()
}
