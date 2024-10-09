package internal

import (
	"context"
	"fmt"
	"github.com/algorandfoundation/hack-tui/api"
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
func (m *StatusModel) Fetch(ctx context.Context, client *api.ClientWithResponses) error {
	s, err := client.GetStatusWithResponse(ctx)
	if err != nil {
		return err
	}
	m.LastRound = uint64(s.JSON200.LastRound)
	return nil
}

// Watch uses algod.StatusAfterBlock to wait for changes and emits to the HeartBeat channel
func (m *StatusModel) Watch(ctx context.Context, client *api.ClientWithResponses) error {
	lastRound := m.LastRound
	for {
		status, err := client.WaitForBlockWithResponse(ctx, int(lastRound))
		if err != nil {
			return err
		}
		m.HeartBeat <- uint64(status.JSON200.LastRound)
		lastRound++
	}
}
