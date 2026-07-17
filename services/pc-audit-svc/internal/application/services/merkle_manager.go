package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type DefaultMerkleManager struct {
	// In reality, this would cache recent hashes and compute the tree.
	// We stub the O(log N) calculation.
}

func NewDefaultMerkleManager() *DefaultMerkleManager {
	return &DefaultMerkleManager{}
}

func (m *DefaultMerkleManager) AddLeaf(ctx context.Context, leafHash string) (string, error) {
	// Stub: Fake tree calculation combining the leaf with an imaginary right node
	h := sha256.New()
	h.Write([]byte(leafHash))
	h.Write([]byte("fake-sibling-hash"))

	newRoot := hex.EncodeToString(h.Sum(nil))
	fmt.Printf("Merkle Tree updated. New Root Hash: %s\n", newRoot)

	return newRoot, nil
}
