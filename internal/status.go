package internal

import (
	"context"
	"fmt"
	"github.com/algorandfoundation/hack-tui/api"
)

// StatusModel represents a status response from algod.Status
type StatusModel struct {
	HeartBeat   chan uint64 // Subscription Channel
	Version     string
	Network     string
	Voting      bool
	NeedsUpdate bool
	LastRound   uint64 // Last recorded round
}

// String prints the last round value
func (m *StatusModel) String() string {
	return fmt.Sprintf("Last round: %d", m.LastRound)
}

// Fetch handles algod.Status
func (m *StatusModel) Fetch(ctx context.Context, client *api.ClientWithResponses) error {
	if m.Version == "" {
		v, err := client.GetVersionWithResponse(ctx)
		if err != nil {
			return err
		}
		if v.StatusCode() != 200 {
			return fmt.Errorf("Satus code %d: %s", v.StatusCode(), v.Status())
		}
		m.Network = v.JSON200.GenesisId
		m.Version = fmt.Sprintf("v%d.%d.%d-%s", v.JSON200.Build.Major, v.JSON200.Build.Minor, v.JSON200.Build.BuildNumber, v.JSON200.Build.Channel)

	}
	m.HeartBeat = make(chan uint64)
	s, err := client.GetStatusWithResponse(ctx)
	if err != nil {
		return err
	}

	if s.StatusCode() != 200 {
		return fmt.Errorf("Satus code %d: %s", s.StatusCode(), s.Status())
	}
	m.LastRound = uint64(s.JSON200.LastRound)

	if s.JSON200.UpgradeNodeVote != nil {
		m.Voting = *s.JSON200.UpgradeNodeVote
	}
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
