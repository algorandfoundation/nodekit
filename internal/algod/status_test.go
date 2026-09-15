package algod

import (
	"context"
	"github.com/algorandfoundation/nodekit/api"
	"github.com/algorandfoundation/nodekit/internal/test"
	"testing"
)

func Test_StatusModel(t *testing.T) {
	m := Status{LastRound: 0}

	emptyCatchpoint := ""

	m = m.Merge(api.StatusLike{LastRound: 5, Catchpoint: &emptyCatchpoint, CatchupTime: 10})
	if m.LastRound != 5 {
		t.Errorf("expected LastRound: 5, got %d", m.LastRound)
	}
	if m.State != SyncingState {
		t.Errorf("expected State: %s, got %s", SyncingState, m.State)
	}

	m = m.Merge(api.StatusLike{LastRound: 10, Catchpoint: &emptyCatchpoint, CatchupTime: 0})
	if m.LastRound != 10 {
		t.Errorf("expected LastRound: 10, got %d", m.LastRound)
	}
	if m.State != StableState {
		t.Errorf("expected State: %s, got %s", StableState, m.State)
	}

	catchpoint := "catchpoint"
	var accountsTotal uint64 = 1000000
	var processedAccounts uint64 = 0
	var verifiedAccounts uint64 = 0
	var keyValueTotal uint64 = 1000
	var keyValueProcessed uint64 = 0
	var keyValueVerified uint64 = 0
	var acquiredBlocks uint64 = 1000000
	var blocksTotal uint64 = 10000000
	m = m.Merge(api.StatusLike{
		LastRound:                   10,
		Catchpoint:                  &catchpoint,
		CatchupTime:                 0,
		CatchpointAcquiredBlocks:    &acquiredBlocks,
		CatchpointTotalBlocks:       &blocksTotal,
		CatchpointTotalAccounts:     &accountsTotal,
		CatchpointVerifiedAccounts:  &verifiedAccounts,
		CatchpointProcessedAccounts: &processedAccounts,
		CatchpointTotalKvs:          &keyValueTotal,
		CatchpointProcessedKvs:      &keyValueProcessed,
		CatchpointVerifiedKvs:       &keyValueVerified,
	})
	if m.State != FastCatchupState {
		t.Errorf("expected State: %s, got %s", FastCatchupState, m.State)
	}

}

func Test_StatusFetch(t *testing.T) {
	client := test.GetClient(true)
	httpPkg := new(api.HttpPkg)

	m, _, err := NewStatus(context.Background(), client, httpPkg)
	if err == nil {
		t.Error("expected error, got nil")
	}

	client = test.NewClient(false, true)
	m, _, err = NewStatus(context.Background(), client, httpPkg)
	if err == nil {
		t.Error("expected error, got nil")
	}

	client = test.GetClient(false)
	m, _, err = NewStatus(context.Background(), client, httpPkg)
	if err != nil {
		t.Error(err)
	}
	if m.LastRound == 0 {
		t.Error("expected LastRound to be non-zero")
	}

}
