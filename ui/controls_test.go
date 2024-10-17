package ui

import (
	"bytes"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"testing"
	"time"
)

func Test_ControlsViewRender(t *testing.T) {
	expected := "(q)uit | (d)elete | (g)enerate | (t)xn | (h)ide"
	// Create the Model
	m := MakeControlViewModel()

	tm := teatest.NewTestModel(
		t, m,
		teatest.WithInitialTermSize(80, 40),
	)

	// Wait for prompt to exit
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte(expected))
		},
		teatest.WithCheckInterval(time.Millisecond*100),
		teatest.WithDuration(time.Second*3),
	)

	// Send quit key
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("q"),
	})

	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
