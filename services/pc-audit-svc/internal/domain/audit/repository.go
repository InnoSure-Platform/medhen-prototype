package audit

import "context"

// HotLedgerRepository manages the real-time, Postgres-backed HEAD.
type HotLedgerRepository interface {
	// AppendLeaf adds a new leaf to the sequential ledger.
	AppendLeaf(ctx context.Context, entry *AuditLedgerEntry) error

	// UpdateMerkleRoot persists the new root hash for the epoch.
	UpdateMerkleRoot(ctx context.Context, epochID int64, rootHash string) error

	// CheckLegalHold returns true if an active hold exists for the given entity or actor.
	CheckLegalHold(ctx context.Context, tenantID, entityID, actorID string) (bool, error)
}

// MerkleTreeManager handles the cryptographic tree math.
type MerkleTreeManager interface {
	// AddLeaf recalculates the Merkle Root when a new leaf is added.
	AddLeaf(ctx context.Context, leafHash string) (newRootHash string, err error)
}

// ColdLakeRepository interacts with Apache Iceberg via REST Catalog.
type ColdLakeRepository interface {
	TimeTravelQuery(ctx context.Context, tenantID string, query string, asOf int64) ([]*AuditLedgerEntry, error)
	InitiateExport(ctx context.Context, job *ExportJob) error
}

// KMSService represents the Key Management System for envelope encryption AND signature verification.
type KMSService interface {
	// VerifySignature executes Zero-Trust non-repudiation checks.
	VerifySignature(ctx context.Context, keyID string, payload []byte, signature []byte) error

	// EncryptPII generates a new DEK and encrypts the payload.
	EncryptPII(ctx context.Context, tenantID string, plaintext []byte) (*CryptoEnvelope, error)

	// DestroyDEK securely shreds the key material.
	DestroyDEK(ctx context.Context, dekReferenceID string) error
}
