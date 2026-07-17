package events

import (
	"encoding/binary"
	"fmt"

	"github.com/hamba/avro/v2"
)

// MagicByte is the Confluent/Apicurio standard byte indicating an Avro payload
const MagicByte byte = 0x0

// SchemaRegistryClient defines the interface for fetching Avro schemas.
type SchemaRegistryClient interface {
	GetSchema(id int) (avro.Schema, error)
}

// AvroSerializer handles serialization of structs into Avro binaries with Schema ID headers.
type AvroSerializer struct {
	registry SchemaRegistryClient
}

// NewAvroSerializer creates a new Avro serializer.
func NewAvroSerializer(registry SchemaRegistryClient) *AvroSerializer {
	return &AvroSerializer{
		registry: registry,
	}
}

// Serialize encodes a payload into Avro format and prepends the 5-byte magic header.
func (s *AvroSerializer) Serialize(schemaID int, payload interface{}) ([]byte, error) {
	schema, err := s.registry.GetSchema(schemaID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schema %d: %w", schemaID, err)
	}

	avroBytes, err := avro.Marshal(schema, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal avro payload: %w", err)
	}

	// 1 magic byte + 4 bytes for schema ID (Big Endian) + avro payload
	msg := make([]byte, 5+len(avroBytes))
	msg[0] = MagicByte
	binary.BigEndian.PutUint32(msg[1:5], uint32(schemaID))
	copy(msg[5:], avroBytes)

	return msg, nil
}
