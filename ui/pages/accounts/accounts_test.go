package accounts

import (
	"bytes"
	"testing"
	"time"

	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/ui/internal/test"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/charmbracelet/x/exp/teatest"
)

func Test_New(t *testing.T) {
	m := New(&algod.StateModel{})
	acc := m.SelectedAccount()

	if acc != nil {
		t.Errorf("Expected no accounts to exist, got %s", acc.Address)
	}
	m, cmd := m.HandleMessage(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("enter"),
	})

	if cmd != nil {
		t.Errorf("Expected no comand")
	}

	m = New(test.GetState(nil))
	m, _ = m.HandleMessage(tea.WindowSizeMsg{Width: 80, Height: 40})

	if m.Data.Admin {
		t.Errorf("Admin flag should be false, got true")
	}

	// Fetch state after message handling
	acc = m.SelectedAccount()
	if acc == nil {
		t.Errorf("expected true, got false")
	}

	// Update syncing state
	m.Data.Status.State = algod.SyncingState
	m.makeRows()
	if m.Data.Status.State != algod.SyncingState {

	}
}

func Test_Snapshot(t *testing.T) {
	t.Run("Visible", func(t *testing.T) {
		model := New(test.GetState(nil))

		model, _ = model.HandleMessage(tea.WindowSizeMsg{Width: 80, Height: 40})
		got := ansi.Strip(model.View())
		golden.RequireEqual(t, []byte(got))
	})
}

func Test_Messages(t *testing.T) {
	// Create the Model
	m := New(test.GetState(nil))

	tm := teatest.NewTestModel(
		t, m,
		teatest.WithInitialTermSize(80, 40),
	)

	// Wait for prompt to exit
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte("accounts | keys"))
		},
		teatest.WithCheckInterval(time.Millisecond*100),
		teatest.WithDuration(time.Second*3),
	)

	tm.Send(*test.GetState(nil))

	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("enter"),
	})

	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("ctrl+c"),
	})

	tm.Send(tea.QuitMsg{})

	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
