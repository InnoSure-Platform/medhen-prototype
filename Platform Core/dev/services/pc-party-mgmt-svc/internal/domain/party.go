package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrDuplicatePartyDetected = errors.New("PTY-1001: duplicate party detected")
	ErrPartyStatusInvalid     = errors.New("PTY-1003: invalid party status for operation")
	ErrPartyAlreadyMerged     = errors.New("PTY-1005: party is already merged")
	ErrInvalidNationalId      = errors.New("PTY-1006: invalid national id")
)

type PartyStatus string

const (
	StatusActive      PartyStatus = "ACTIVE"
	StatusSuspended   PartyStatus = "SUSPENDED"
	StatusBlacklisted PartyStatus = "BLACKLISTED"
	StatusMerged      PartyStatus = "MERGED"
	StatusAnonymized  PartyStatus = "ANONYMIZED"
)

type PartyType string

const (
	TypeIndividual   PartyType = "INDIVIDUAL"
	TypeOrganization PartyType = "ORGANIZATION"
)

type Party struct {
	ID                uuid.UUID
	TenantID          string
	Type              PartyType
	Status            PartyStatus
	KYCStatus         KYCStatus
	FirstName         string
	LastName          string
	DOB               *time.Time
	Gender            string
	NationalIDType    string
	NationalIDNumber  string
	LegalName         string
	RegistrationNo    string
	IndustryCode      string
	TIN               string
	SurvivingPartyID  *uuid.UUID
	Version           int
	CreatedAt         time.Time
	UpdatedAt         time.Time
	Addresses         []Address
	Roles             []PartyRole
	BankAccounts      []BankAccount
}

type BankAccount struct {
	ID            uuid.UUID
	PartyID       uuid.UUID
	BankCode      string
	AccountNumber string // Encrypted at rest
	IsPrimary     bool
}

func NewIndividual(tenantID, firstName, lastName string, dob time.Time, gender, idType, idNum, tin string) (*Party, error) {
	if idNum == "" {
		return nil, ErrInvalidNationalId
	}
	return &Party{
		ID:               uuid.New(),
		TenantID:         tenantID,
		Type:             TypeIndividual,
		Status:           StatusActive,
		KYCStatus:        KYCStatusPending,
		FirstName:        firstName,
		LastName:         lastName,
		DOB:              &dob,
		Gender:           gender,
		NationalIDType:   idType,
		NationalIDNumber: idNum,
		TIN:              tin,
		Version:          1,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}, nil
}

func NewOrganization(tenantID, legalName, regNo, indCode, tin string) (*Party, error) {
	if regNo == "" {
		return nil, ErrInvalidNationalId
	}
	return &Party{
		ID:             uuid.New(),
		TenantID:       tenantID,
		Type:           TypeOrganization,
		Status:         StatusActive,
		KYCStatus:      KYCStatusPending,
		LegalName:      legalName,
		RegistrationNo: regNo,
		IndustryCode:   indCode,
		TIN:            tin,
		Version:        1,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}, nil
}

func (p *Party) MergeInto(targetID uuid.UUID) error {
	if p.Status == StatusMerged {
		return ErrPartyAlreadyMerged
	}
	if p.Status == StatusBlacklisted {
		return ErrPartyStatusInvalid
	}
	
	p.Status = StatusMerged
	p.SurvivingPartyID = &targetID
	p.UpdatedAt = time.Now()
	p.Version++
	return nil
}

func (p *Party) Suspend() error {
	if p.Status != StatusActive {
		return ErrPartyStatusInvalid
	}
	p.Status = StatusSuspended
	p.UpdatedAt = time.Now()
	p.Version++
	return nil
}

type PartyRole struct {
	ID            uuid.UUID
	PartyID       uuid.UUID
	Role          string
	Attributes    map[string]interface{}
	EffectiveFrom time.Time
	EffectiveTo   *time.Time
}
