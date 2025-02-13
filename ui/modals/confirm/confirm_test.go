package confirm

import (
	"bytes"
	"github.com/algorandfoundation/nodekit/internal/test/mock"
	"github.com/algorandfoundation/nodekit/ui/internal/test"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/charmbracelet/x/exp/teatest"
	"testing"
	"time"
)

func Test_New(t *testing.T) {
	m := New(test.GetState(nil), nil)
	if m.Participation != nil {
		t.Errorf("expected ActiveKey to be nil")
	}
	m.Participation = &mock.Keys[0]
	// Handle Delete
	m, cmd := m.HandleMessage(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("y"),
	})

	if cmd == nil {
		t.Errorf("expected cmd to be non-nil")
	}
}
func Test_Snapshot(t *testing.T) {
	t.Run("NoKey", func(t *testing.T) {
		model := New(test.GetState(nil), nil)
		got := ansi.Strip(model.View())
		golden.RequireEqual(t, []byte(got))
	})
	t.Run("Visible", func(t *testing.T) {
		model := New(test.GetState(nil), &mock.Keys[0])
		model, _ = model.HandleMessage(tea.WindowSizeMsg{Width: 80, Height: 40})
		got := ansi.Strip(model.View())
		golden.RequireEqual(t, []byte(got))
	})
}

func Test_Messages(t *testing.T) {
	// Create the Model
	m := New(test.GetState(nil), &mock.Keys[0])
	tm := teatest.NewTestModel(
		t, m,
		teatest.WithInitialTermSize(80, 40),
	)

	// Wait for prompt to exit
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte("Are you sure you want to delete this key from your node?"))
		},
		teatest.WithCheckInterval(time.Millisecond*100),
		teatest.WithDuration(time.Second*3),
	)

	tm.Send(*test.GetState(nil))

	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("n"),
	})
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("esc"),
	})

	tm.Send(tea.QuitMsg{})

	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
