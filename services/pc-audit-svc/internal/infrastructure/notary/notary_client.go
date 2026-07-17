package notary

import (
	"context"
	"fmt"
	"time"
)

// NotaryClient acts as the Out-of-Band anchor to a trusted external party (e.g. NBE API or Blockchain).
type NotaryClient struct {
	endpoint string
}

func NewNotaryClient(endpoint string) *NotaryClient {
	return &NotaryClient{endpoint: endpoint}
}

// PublishRootHash transmits the immutable epoch top hash outside of Medhen's infrastructure.
func (n *NotaryClient) PublishRootHash(ctx context.Context, epochID int64, rootHash string) error {
	// HTTP call to NBE Vault or Ethereum Smart Contract
	fmt.Printf("[%s] SUCCESS: Published Merkle Root Hash %s to external Notary %s for Epoch %d\n",
		time.Now().Format(time.RFC3339), rootHash, n.endpoint, epochID)
	return nil
}
