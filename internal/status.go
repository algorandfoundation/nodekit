package internal

import (
	"context"
	"fmt"
	"github.com/algorand/go-algorand-sdk/client/v2/algod"
)

// StatusModel represents a status response from algod.Status
type StatusModel struct {
	LastRound int // Last recorded round
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
	m.LastRound = int(s.LastRound)
	return nil
}
