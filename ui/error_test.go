package ui

import (
	"bytes"
	"testing"
	"time"

	"github.com/algorandfoundation/hack-tui/ui/controls"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/charmbracelet/x/exp/teatest"
)

func Test_ErrorSnapshot(t *testing.T) {
	t.Run("Visible", func(t *testing.T) {
		model := ErrorViewModel{
			Height:   20,
			Width:    40,
			controls: controls.New(" Error "),
			Message:  "a test error",
		}
		got := ansi.Strip(model.View())
		golden.RequireEqual(t, []byte(got))
	})
}

func Test_ErrorMessages(t *testing.T) {
	tm := teatest.NewTestModel(
		t, ErrorViewModel{Message: "a test error"},
		teatest.WithInitialTermSize(120, 80),
	)

	// Wait for prompt to exit
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte("a test error"))
		},
		teatest.WithCheckInterval(time.Millisecond*100),
		teatest.WithDuration(time.Second*3),
	)
	// Resize Message
	tm.Send(tea.WindowSizeMsg{
		Width:  50,
		Height: 20,
	})

	// Send quit key
	tm.Send(tea.QuitMsg{})

	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
