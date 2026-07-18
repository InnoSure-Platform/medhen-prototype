// Package domain is the party bounded context: individuals and organizations
// (policyholders, claimants, intermediaries) and their Ethiopian addresses. It
// is pure — no framework, DB, or HTTP — and depends only on the platform kernel.
package domain

import (
	"errors"
	"time"

	"github.com/InnoSure-Platform/medhen-prototype/internal/platform/ids"
)

// Type is the kind of party.
type Type string

const (
	Individual   Type = "INDIVIDUAL"
	Organization Type = "ORGANIZATION"
)

// Status is the party lifecycle state.
type Status string

const (
	StatusActive     Status = "ACTIVE"
	StatusSuspended  Status = "SUSPENDED"
	StatusAnonymized Status = "ANONYMIZED"
)

var (
	ErrNameRequired       = errors.New("party: name is required")
	ErrPhoneRequired      = errors.New("party: phone (E.164) is required")
	ErrNationalIDRequired = errors.New("party: national id is required")
	ErrInvalidAddress     = errors.New("party: region, zone and woreda are required")
	ErrInvalidTransition  = errors.New("party: invalid status transition")
)

// Address is an Ethiopian administrative address (Region → Zone → Woreda → Kebele).
type Address struct {
	Region      string
	Zone        string
	Woreda      string
	Kebele      string
	HouseNumber string
}

func (a Address) validate() error {
	if a.Region == "" || a.Zone == "" || a.Woreda == "" {
		return ErrInvalidAddress
	}
	return nil
}

// Party is the aggregate root.
type Party struct {
	ID              string
	TenantID        string
	Type            Type
	Status          Status
	FullName        string
	FullNameAmharic string
	PhoneE164       string
	NationalID      string
	Address         Address
	Version         int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewIndividual creates an active individual party after validating inputs.
func NewIndividual(tenantID, fullName, fullNameAmharic, phone, nationalID string, addr Address) (*Party, error) {
	if fullName == "" {
		return nil, ErrNameRequired
	}
	if phone == "" {
		return nil, ErrPhoneRequired
	}
	if nationalID == "" {
		return nil, ErrNationalIDRequired
	}
	if err := addr.validate(); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	return &Party{
		ID:              ids.New(),
		TenantID:        tenantID,
		Type:            Individual,
		Status:          StatusActive,
		FullName:        fullName,
		FullNameAmharic: fullNameAmharic,
		PhoneE164:       phone,
		NationalID:      nationalID,
		Address:         addr,
		Version:         1,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// NewOrganization creates an active organization party.
func NewOrganization(tenantID, legalName, legalNameAmharic, phone, registrationNo string, addr Address) (*Party, error) {
	if legalName == "" {
		return nil, ErrNameRequired
	}
	if registrationNo == "" {
		return nil, ErrNationalIDRequired
	}
	if err := addr.validate(); err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	return &Party{
		ID:              ids.New(),
		TenantID:        tenantID,
		Type:            Organization,
		Status:          StatusActive,
		FullName:        legalName,
		FullNameAmharic: legalNameAmharic,
		PhoneE164:       phone,
		NationalID:      registrationNo,
		Address:         addr,
		Version:         1,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// Suspend transitions an active party to suspended.
func (p *Party) Suspend() error {
	if p.Status != StatusActive {
		return ErrInvalidTransition
	}
	p.Status = StatusSuspended
	p.touch()
	return nil
}

func (p *Party) touch() {
	p.UpdatedAt = time.Now().UTC()
	p.Version++
}
