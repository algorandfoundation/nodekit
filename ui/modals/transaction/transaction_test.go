package transaction

import (
	"bytes"
	"github.com/algorandfoundation/hack-tui/test"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/charmbracelet/x/exp/teatest"
	"testing"
	"time"
)

func Test_New(t *testing.T) {
	model := New(test.GetState(nil))
	model.Participation = &test.Keys[0]
	model.Participation.Address = "ALGO123456789"
	addr := model.FormatedAddress()
	if addr != "ALGO...6789" {
		t.Errorf("Expected ALGO123456789, got %s", addr)
	}
	model.Participation.Address = "ABC"
}
func Test_Snapshot(t *testing.T) {
	t.Run("NotVisible", func(t *testing.T) {
		model := New(test.GetState(nil))
		model.Participation = &test.Keys[0]
		model.UpdateState()
		got := ansi.Strip(model.View())
		golden.RequireEqual(t, []byte(got))
	})
	t.Run("Offline", func(t *testing.T) {
		model := New(test.GetState(nil))
		model.Participation = &test.Keys[0]
		model, _ = model.HandleMessage(tea.WindowSizeMsg{
			Height: 40,
			Width:  80,
		})
		model.Active = true
		model.UpdateState()
		got := ansi.Strip(model.View())
		golden.RequireEqual(t, []byte(got))
	})
	t.Run("Online", func(t *testing.T) {
		model := New(test.GetState(nil))
		model.Participation = &test.Keys[0]
		model, _ = model.HandleMessage(tea.WindowSizeMsg{
			Height: 40,
			Width:  80,
		})
		model.UpdateState()
		got := ansi.Strip(model.View())
		golden.RequireEqual(t, []byte(got))
	})

	t.Run("Loading", func(t *testing.T) {
		model := New(test.GetState(nil))
		model.Participation = &test.Keys[0]
		got := ansi.Strip(model.View())
		golden.RequireEqual(t, []byte(got))
	})
	t.Run("NoKey", func(t *testing.T) {
		model := New(test.GetState(nil))
		got := ansi.Strip(model.View())
		golden.RequireEqual(t, []byte(got))
	})
}

func Test_Messages(t *testing.T) {
	// Create the Model
	m := New(test.GetState(nil))
	m.Participation = &test.Keys[0]

	tm := teatest.NewTestModel(
		t, m,
		teatest.WithInitialTermSize(80, 40),
	)

	// Wait for prompt to exit
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte("████████"))
		},
		teatest.WithCheckInterval(time.Millisecond*100),
		teatest.WithDuration(time.Second*3),
	)

	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("esc"),
	})

	tm.Send(tea.QuitMsg{})

	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
