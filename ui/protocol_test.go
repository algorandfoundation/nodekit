package ui

import (
	"bytes"
	"testing"
	"time"

	"github.com/algorandfoundation/hack-tui/internal"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/charmbracelet/x/exp/teatest"
)

var protocolViewSnapshots = map[string]ProtocolViewModel{
	"Hidden": {
		Data: internal.StatusModel{
			State:       "SYNCING",
			Version:     "v0.0.0-test",
			Network:     "test-v1",
			Voting:      true,
			NeedsUpdate: true,
			LastRound:   0,
		},
		TerminalWidth:  60,
		TerminalHeight: 40,
		IsVisible:      false,
	},
	"HiddenHeight": {
		Data: internal.StatusModel{
			State:       "SYNCING",
			Version:     "v0.0.0-test",
			Network:     "test-v1",
			Voting:      true,
			NeedsUpdate: true,
			LastRound:   0,
		},
		TerminalWidth:  70,
		TerminalHeight: 20,
		IsVisible:      true,
	},
	"Visible": {
		Data: internal.StatusModel{
			State:       "SYNCING",
			Version:     "v0.0.0-test",
			Network:     "test-v1",
			Voting:      true,
			NeedsUpdate: true,
			LastRound:   0,
		},
		TerminalWidth:  160,
		TerminalHeight: 80,
		IsVisible:      true,
	},
	"VisibleSmall": {
		Data: internal.StatusModel{
			State:       "SYNCING",
			Version:     "v0.0.0-test",
			Network:     "test-v1",
			Voting:      true,
			NeedsUpdate: true,
			LastRound:   0,
		},
		TerminalWidth:  80,
		TerminalHeight: 40,
		IsVisible:      true,
	},
	"NoVoteOrUpgrade": {
		Data: internal.StatusModel{
			State:       "SYNCING",
			Version:     "v0.0.0-test",
			Network:     "test-v1",
			Voting:      false,
			NeedsUpdate: false,
			LastRound:   0,
		},
		TerminalWidth:  160,
		TerminalHeight: 80,
		IsVisible:      true,
	},
	"NoVoteOrUpgradeSmall": {
		Data: internal.StatusModel{
			State:       "SYNCING",
			Version:     "v0.0.0-test",
			Network:     "test-v1",
			Voting:      false,
			NeedsUpdate: false,
			LastRound:   0,
		},
		TerminalWidth:  80,
		TerminalHeight: 40,
		IsVisible:      true,
	},
}

func Test_ProtocolSnapshot(t *testing.T) {
	for name, model := range protocolViewSnapshots {
		t.Run(name, func(t *testing.T) {
			got := ansi.Strip(model.View())
			golden.RequireEqual(t, []byte(got))
		})
	}
}

// Test_ProtocolMessages handles any additional tests like sending messages
func Test_ProtocolMessages(t *testing.T) {
	state := internal.StateModel{
		Status: internal.StatusModel{
			LastRound:   1337,
			NeedsUpdate: true,
			State:       internal.SyncingState,
		},
		Metrics: internal.MetricsModel{
			RoundTime: 0,
			TX:        0,
			RX:        0,
			TPS:       0,
		},
	}

	// Create the Model
	m := MakeProtocolViewModel(&state)

	tm := teatest.NewTestModel(
		t, m,
		teatest.WithInitialTermSize(120, 80),
	)

	// Wait for prompt to exit
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte("[UPDATE AVAILABLE]"))
		},
		teatest.WithCheckInterval(time.Millisecond*100),
		teatest.WithDuration(time.Second*3),
	)
	tm.Send(internal.StatusModel{
		State:       "",
		Version:     "",
		Network:     "",
		Voting:      false,
		NeedsUpdate: false,
		LastRound:   0,
	})
	// Send hide key
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("h"),
	})

	// Send quit key
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("ctrl+c"),
	})

	// Send quit msg
	tm.Send(tea.QuitMsg{})

	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
