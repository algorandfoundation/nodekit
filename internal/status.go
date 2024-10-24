package internal

import (
	"context"
	"fmt"
	"github.com/algorandfoundation/hack-tui/api"
)

// StatusModel represents a status response from algod.Status
type StatusModel struct {
	State       string
	Version     string
	Network     string
	Voting      bool
	NeedsUpdate bool
	LastRound   uint64 // Last recorded round
}

// String prints the last round value
func (m *StatusModel) String() string {
	return fmt.Sprintf("\nLastRound: %d\n", m.LastRound)
}

// Fetch handles algod.Status
func (m *StatusModel) Fetch(ctx context.Context, client *api.ClientWithResponses) error {
	if m.Version == "" || m.Version == "NA" {
		v, err := client.GetVersionWithResponse(ctx)
		if err != nil {
			return err
		}
		if v.StatusCode() != 200 {
			return fmt.Errorf("Status code %d: %s", v.StatusCode(), v.Status())
		}
		m.Network = v.JSON200.GenesisId
		m.Version = fmt.Sprintf("v%d.%d.%d-%s", v.JSON200.Build.Major, v.JSON200.Build.Minor, v.JSON200.Build.BuildNumber, v.JSON200.Build.Channel)

	}

	s, err := client.GetStatusWithResponse(ctx)
	if err != nil {
		return err
	}

	if s.StatusCode() != 200 {
		return fmt.Errorf("Status code %d: %s", s.StatusCode(), s.Status())
	}
	m.LastRound = uint64(s.JSON200.LastRound)

	if s.JSON200.UpgradeNodeVote != nil {
		m.Voting = *s.JSON200.UpgradeNodeVote
	}
	return nil
}
