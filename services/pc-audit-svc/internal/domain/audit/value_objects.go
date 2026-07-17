package audit

// ActorContext represents the identity and context of the user making the mutation.
type ActorContext struct {
	UserID    string
	Role      string
	IPAddress string
}

// CryptoEnvelope holds the encrypted PII payload and its associated KMS Key reference.
// Destroying the DEK in the KMS permanently shreds the data (Right to be Forgotten).
type CryptoEnvelope struct {
	DEKReferenceID string
	Ciphertext     []byte
}

// StateDelta represents the JSON diff of the before and after states.
type StateDelta struct {
	Before []byte
	After  []byte
}
