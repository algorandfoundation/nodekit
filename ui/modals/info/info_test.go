package info

import (
	"bytes"
	"github.com/algorandfoundation/hack-tui/internal/test/mock"
	"github.com/algorandfoundation/hack-tui/ui/internal/test"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/charmbracelet/x/exp/teatest"
	"testing"
	"time"
)

func Test_New(t *testing.T) {
	m := New(test.GetState(nil))
	if m == nil {
		t.Fatal("New returned nil")
	}
	m.Participation = &mock.Keys[0]
	account := m.State.Accounts[mock.Keys[0].Address]
	account.Status = "Online"
	m.State.Accounts[mock.Keys[0].Address] = account
	m.Active = true
	m.UpdateState()
	if m.BorderColor != "1" {
		t.Error("State is not correct, border should be 1")
	}
	if m.Controls != "( take (o)ffline )" {
		t.Error("Controls are not correct")
	}
}
func Test_Snapshot(t *testing.T) {
	t.Run("Visible", func(t *testing.T) {
		model := New(test.GetState(nil))
		model.Participation = &mock.Keys[0]
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
	m.Participation = &mock.Keys[0]

	tm := teatest.NewTestModel(
		t, m,
		teatest.WithInitialTermSize(80, 40),
	)

	// Wait for prompt to exit
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte("Account: ABC"))
		},
		teatest.WithCheckInterval(time.Millisecond*100),
		teatest.WithDuration(time.Second*3),
	)

	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("o"),
	})
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("d"),
	})
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("esc"),
	})

	tm.Send(tea.QuitMsg{})

	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
