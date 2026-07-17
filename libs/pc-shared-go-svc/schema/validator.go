package schema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// Validator provides a thread-safe registry of compiled JSON-Schemas.
type Validator struct {
	mu       sync.RWMutex
	compiler *jsonschema.Compiler
	schemas  map[string]*jsonschema.Schema
}

// NewValidator initializes a new schema validator engine.
func NewValidator() *Validator {
	return &Validator{
		compiler: jsonschema.NewCompiler(),
		schemas:  make(map[string]*jsonschema.Schema),
	}
}

// Register compiles and caches a JSON-Schema under a unique ID.
func (v *Validator) Register(schemaID string, schemaPayload []byte) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if err := v.compiler.AddResource(schemaID, bytes.NewReader(schemaPayload)); err != nil {
		return fmt.Errorf("failed to add schema resource %s: %w", schemaID, err)
	}

	compiled, err := v.compiler.Compile(schemaID)
	if err != nil {
		return fmt.Errorf("failed to compile schema %s: %w", schemaID, err)
	}

	v.schemas[schemaID] = compiled
	return nil
}

// Validate checks a JSON payload against a registered schema ID.
func (v *Validator) Validate(schemaID string, payload []byte) error {
	v.mu.RLock()
	schema, exists := v.schemas[schemaID]
	v.mu.RUnlock()

	if !exists {
		return fmt.Errorf("schema %s not found in registry", schemaID)
	}

	var parsed interface{}
	if err := json.Unmarshal(payload, &parsed); err != nil {
		return fmt.Errorf("invalid json payload: %w", err)
	}

	if err := schema.Validate(parsed); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}
