package domain

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrInvalidAdminUnitMapping = errors.New("PTY-1002: invalid admin unit mapping")
)

type AddressType string

const (
	AddressTypeResidential AddressType = "RESIDENTIAL"
	AddressTypeMailing     AddressType = "MAILING"
	AddressTypeBusiness    AddressType = "BUSINESS"
)

type Address struct {
	ID          uuid.UUID
	PartyID     uuid.UUID
	Type        AddressType
	IsPrimary   bool
	Region      string
	Zone        string
	Woreda      string
	Kebele      string
	HouseNumber string
	Version     int
}

func NewAddress(partyID uuid.UUID, addrType AddressType, region, zone, woreda, kebele, houseNumber string, isPrimary bool) (*Address, error) {
	if region == "" || zone == "" || woreda == "" {
		return nil, ErrInvalidAdminUnitMapping
	}
	
	return &Address{
		ID:          uuid.New(),
		PartyID:     partyID,
		Type:        addrType,
		IsPrimary:   isPrimary,
		Region:      region,
		Zone:        zone,
		Woreda:      woreda,
		Kebele:      kebele,
		HouseNumber: houseNumber,
		Version:     1,
	}, nil
}
