package internal

import (
	"context"
	"fmt"
	"github.com/algorand/go-algorand-sdk/client/v2/algod"
)

// StatusModel represents a status response from algod.Status
type StatusModel struct {
	HeartBeat chan uint64 // Subscription Channel
	LastRound uint64      // Last recorded round
}

// String prints the last round value
func (m *StatusModel) String() string {
	return fmt.Sprintf("Last round: %d", m.LastRound)
}

// Fetch handles algod.Status
func (m *StatusModel) Fetch(algodClient *algod.Client) error {
	s, err := algodClient.Status().Do(context.Background())
	if err != nil {
		return err
	}
	m.LastRound = s.LastRound
	return nil
}

// Watch uses algod.StatusAfterBlock to wait for changes and emits to the HeartBeat channel
func (m *StatusModel) Watch(ctx context.Context, algodClient *algod.Client) error {
	lastRound := m.LastRound
	for {
		status, err := algodClient.StatusAfterBlock(lastRound).Do(ctx)
		if err != nil {
			return err
		}
		m.HeartBeat <- status.LastRound
		lastRound++
	}
}
