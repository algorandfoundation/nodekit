package exception

import (
	"bytes"
	"errors"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/charmbracelet/x/exp/teatest"
	"testing"
	"time"
)

func Test_Snapshot(t *testing.T) {
	t.Run("Visible", func(t *testing.T) {
		model := New("Something went wrong")
		got := ansi.Strip(model.View())
		golden.RequireEqual(t, []byte(got))
	})
}

func Test_Messages(t *testing.T) {
	// Create the Model
	m := New("Something went wrong")
	tm := teatest.NewTestModel(
		t, m,
		teatest.WithInitialTermSize(80, 40),
	)

	// Wait for prompt to exit
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte("Something went wrong"))
		},
		teatest.WithCheckInterval(time.Millisecond*100),
		teatest.WithDuration(time.Second*3),
	)

	tm.Send(errors.New("Something else went wrong"))
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("esc"),
	})

	tm.Send(tea.QuitMsg{})

	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
