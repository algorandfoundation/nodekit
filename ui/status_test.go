package ui

import (
	"bytes"
	"github.com/algorandfoundation/hack-tui/internal"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/charmbracelet/x/exp/teatest"
	"testing"
	"time"
)

var statusViewSnapshots = map[string]StatusViewModel{
	"Syncing": {
		Data: &internal.StateModel{
			Status: internal.StatusModel{
				LastRound:   1337,
				NeedsUpdate: true,
				State:       "SYNCING",
			},
			Metrics: internal.MetricsModel{
				RoundTime: 0,
				TX:        0,
			},
		},
		TerminalWidth:  180,
		TerminalHeight: 80,
		IsVisible:      true,
	},
	"Hidden": {
		Data: &internal.StateModel{
			Status: internal.StatusModel{
				LastRound:   1337,
				NeedsUpdate: true,
				State:       "SYNCING",
			},
			Metrics: internal.MetricsModel{
				RoundTime: 0,
				TX:        0,
			},
		},
		TerminalWidth:  180,
		TerminalHeight: 80,
		IsVisible:      false,
	},
	"Loading": {
		Data: &internal.StateModel{
			Status: internal.StatusModel{
				LastRound:   1337,
				NeedsUpdate: true,
				State:       "SYNCING",
			},
			Metrics: internal.MetricsModel{
				RoundTime: 0,
				TX:        0,
			},
		},
		TerminalWidth:  0,
		TerminalHeight: 0,
		IsVisible:      true,
	},
}

func Test_StatusSnapshot(t *testing.T) {
	for name, model := range statusViewSnapshots {
		t.Run(name, func(t *testing.T) {
			got := ansi.Strip(model.View())
			golden.RequireEqual(t, []byte(got))
		})
	}
}

func Test_StatusMessages(t *testing.T) {
	state := internal.StateModel{
		Status: internal.StatusModel{
			LastRound:   1337,
			NeedsUpdate: true,
			State:       "SYNCING",
		},
		Metrics: internal.MetricsModel{
			RoundTime: 0,
			TX:        0,
			RX:        0,
			TPS:       0,
		},
	}
	// Create the Model
	m := StatusViewModel{
		Data:          &state,
		TerminalWidth: 80,
		IsVisible:     true,
	}

	tm := teatest.NewTestModel(
		t, m,
		teatest.WithInitialTermSize(80, 40),
	)

	// Wait for prompt to exit
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte("Latest Round: 1337"))
		},
		teatest.WithCheckInterval(time.Millisecond*100),
		teatest.WithDuration(time.Second*3),
	)

	// Send the state
	tm.Send(state)

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

	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
