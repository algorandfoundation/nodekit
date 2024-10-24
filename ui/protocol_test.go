package ui

import (
	"bytes"
	"github.com/algorandfoundation/hack-tui/internal"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"testing"
	"time"
)

func Test_ProtocolViewRender(t *testing.T) {
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
