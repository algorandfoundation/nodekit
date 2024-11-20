package internal

import (
	"context"
	"github.com/algorandfoundation/hack-tui/internal/test"
	"strings"
	"testing"
)

func Test_StatusModel(t *testing.T) {
	m := StatusModel{LastRound: 0}
	if !strings.Contains(m.String(), "LastRound: 0") {
		t.Fatal("expected \"LastRound: 0\", got ", m.String())
	}

	stale := true
	m.Update(5, 10, &stale)

	if m.LastRound != 5 {
		t.Errorf("expected LastRound: 5, got %d", m.LastRound)
	}
	if m.State != SyncingState {
		t.Errorf("expected State: %s, got %s", SyncingState, m.State)
	}

	m.Update(10, 0, &stale)
	if m.LastRound != 10 {
		t.Errorf("expected LastRound: 10, got %d", m.LastRound)
	}
	if m.State != StableState {
		t.Errorf("expected State: %s, got %s", StableState, m.State)
	}

}

func Test_StatusFetch(t *testing.T) {
	m := StatusModel{LastRound: 0}
	err := m.Fetch(context.Background(), test.GetClient(false))
	if err != nil {
		t.Error(err)
	}
	if m.LastRound == 0 {
		t.Error("expected LastRound to be non-zero")
	}
}
