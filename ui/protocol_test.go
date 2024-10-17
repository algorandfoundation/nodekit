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
	status := internal.StatusModel{
		HeartBeat:   make(chan uint64),
		LastRound:   0,
		NeedsUpdate: true,
		State:       "SYNCING",
	}

	// Create the Model
	m := MakeProtocolViewModel(&status)

	tm := teatest.NewTestModel(
		t, m,
		teatest.WithInitialTermSize(80, 40),
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
		Runes: []rune("q"),
	})

	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
