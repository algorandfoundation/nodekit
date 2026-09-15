package statewalker

import (
	"context"
	"fmt"
	"time"

	"github.com/algorandfoundation/nodekit/api"
	"github.com/algorandfoundation/nodekit/internal/algod"
)

// DefaultPollInterval is the default cadence for polling /v2/status.
const DefaultPollInterval = time.Second

// ExpectedState describes the algod state a scenario expects nodekit to observe.
type ExpectedState struct {
	// State is the operational state derived from /v2/status,
	// one of algod.SyncingState, algod.FastCatchupState or algod.StableState.
	State algod.State

	// MinRound, when non-zero, requires LastRound to be at least this value.
	MinRound uint64

	// UpgradeVoting, when true, requires UpgradeVoteRounds to be non-zero.
	UpgradeVoting bool
}

// Matches reports whether the given status satisfies the expectation.
func (e ExpectedState) Matches(status algod.Status) bool {
	if e.State != "" && status.State != e.State {
		return false
	}
	if e.MinRound > 0 && status.LastRound < e.MinRound {
		return false
	}
	if e.UpgradeVoting && status.UpgradeVoteRounds == 0 {
		return false
	}
	return true
}

// WaitFor polls /v2/status through the given client until the expected state
// is observed or the timeout elapses. It reuses nodekit's algod.Status mapping
// so scenarios assert exactly what the TUI would render.
func WaitFor(ctx context.Context, client api.ClientWithResponsesInterface, expected ExpectedState, timeout time.Duration) error {
	return waitFor(ctx, client, expected, timeout, DefaultPollInterval)
}

func waitFor(ctx context.Context, client api.ClientWithResponsesInterface, expected ExpectedState, timeout time.Duration, interval time.Duration) error {
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()

	var lastErr error
	var lastStatus algod.Status
	for {
		status, _, err := algod.Status{Client: client}.Get(ctx)
		lastErr = err
		if err == nil {
			lastStatus = status
			if expected.Matches(status) {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline.C:
			if lastErr != nil {
				return fmt.Errorf("timed out waiting for state %+v: last error: %w", expected, lastErr)
			}
			return fmt.Errorf("timed out waiting for state %+v: last observed state %s at round %d", expected, lastStatus.State, lastStatus.LastRound)
		case <-time.After(interval):
		}
	}
}

// CurrentStatus fetches the node status through nodekit's status mapping, so
// callers observe exactly what the TUI would render.
func CurrentStatus(ctx context.Context, client api.ClientWithResponsesInterface) (algod.Status, api.ResponseInterface, error) {
	return algod.Status{Client: client}.Get(ctx)
}

// LastCatchpoint returns the most recent catchpoint label seen by the node,
// or an empty string when no catchpoint is available yet.
func LastCatchpoint(ctx context.Context, client api.ClientWithResponsesInterface) (string, error) {
	response, err := client.GetStatusWithResponse(ctx)
	if err != nil {
		return "", err
	}
	if response.StatusCode() != 200 || response.JSON200 == nil {
		return "", fmt.Errorf("failed to get status: %s", response.Status())
	}
	if response.JSON200.LastCatchpoint == nil {
		return "", nil
	}
	return *response.JSON200.LastCatchpoint, nil
}

// WaitForRoundsToAdvance observes /v2/status until LastRound has increased by
// at least rounds, proving the chain is making progress without traffic.
func WaitForRoundsToAdvance(ctx context.Context, client api.ClientWithResponsesInterface, rounds uint64, timeout time.Duration) error {
	status, _, err := algod.Status{Client: client}.Get(ctx)
	if err != nil {
		return err
	}
	return waitFor(ctx, client, ExpectedState{MinRound: status.LastRound + rounds}, timeout, DefaultPollInterval)
}
