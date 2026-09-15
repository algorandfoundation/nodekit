package statewalker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/algorandfoundation/nodekit/api"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/internal/test"
)

// errClient simulates a node that is unreachable.
type errClient struct {
	api.ClientWithResponsesInterface
}

func (c *errClient) GetStatusWithResponse(ctx context.Context, reqEditors ...api.RequestEditorFn) (*api.GetStatusResponse, error) {
	return nil, errors.New("connection refused")
}

func Test_ExpectedState_Matches(t *testing.T) {
	status := algod.Status{
		State:             algod.StableState,
		LastRound:         100,
		UpgradeVoteRounds: 0,
	}

	if !(ExpectedState{}).Matches(status) {
		t.Error("empty expectation should match any status")
	}
	if !(ExpectedState{State: algod.StableState}).Matches(status) {
		t.Error("expected RUNNING to match")
	}
	if (ExpectedState{State: algod.SyncingState}).Matches(status) {
		t.Error("expected SYNCING not to match a RUNNING status")
	}
	if !(ExpectedState{MinRound: 100}).Matches(status) {
		t.Error("expected MinRound at the current round to match")
	}
	if (ExpectedState{MinRound: 101}).Matches(status) {
		t.Error("expected MinRound above the current round not to match")
	}
	if (ExpectedState{UpgradeVoting: true}).Matches(status) {
		t.Error("expected UpgradeVoting not to match without vote rounds")
	}
	status.UpgradeVoteRounds = 10000
	if !(ExpectedState{UpgradeVoting: true}).Matches(status) {
		t.Error("expected UpgradeVoting to match with vote rounds")
	}
}

func Test_WaitFor(t *testing.T) {
	ctx := context.Background()
	client := test.GetClient(false)

	// The mock client reports a stable node at round 10
	err := WaitFor(ctx, client, ExpectedState{State: algod.StableState, MinRound: 10}, time.Second)
	if err != nil {
		t.Errorf("expected stable state to be observed, got: %s", err)
	}

	// An unreachable round should time out with the last observed state
	err = waitFor(ctx, client, ExpectedState{MinRound: 11}, 50*time.Millisecond, 10*time.Millisecond)
	if err == nil {
		t.Error("expected a timeout waiting for an unreachable round")
	}

	// An unreachable node should time out with the last error
	err = waitFor(ctx, &errClient{}, ExpectedState{State: algod.StableState}, 50*time.Millisecond, 10*time.Millisecond)
	if err == nil {
		t.Error("expected a timeout for an unreachable node")
	}

	// A cancelled context should abort the wait
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	err = waitFor(cancelled, client, ExpectedState{MinRound: 11}, time.Second, 10*time.Millisecond)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got: %v", err)
	}
}

func Test_WaitForRoundsToAdvance(t *testing.T) {
	ctx := context.Background()
	client := test.GetClient(false)

	// The mock client never advances past round 10, so expect a timeout
	err := WaitForRoundsToAdvance(ctx, client, 1, 50*time.Millisecond)
	if err == nil {
		t.Error("expected a timeout when rounds do not advance")
	}

	// An unreachable node should surface the error immediately
	err = WaitForRoundsToAdvance(ctx, &errClient{}, 1, 50*time.Millisecond)
	if err == nil {
		t.Error("expected an error for an unreachable node")
	}
}
