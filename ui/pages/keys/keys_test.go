package keys

import (
	"bytes"
	"github.com/algorandfoundation/algorun-tui/api"
	"github.com/algorandfoundation/algorun-tui/internal/test/mock"
	"github.com/algorandfoundation/algorun-tui/ui/app"
	"github.com/algorandfoundation/algorun-tui/ui/internal/test"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/charmbracelet/x/exp/teatest"
	"testing"
	"time"
)

func Test_New(t *testing.T) {
	m := New("ABC", nil)
	if m.Address != "ABC" {
		t.Errorf("Expected Address to be ABC, got %s", m.Address)
	}
	d, active := m.SelectedKey()
	if active {
		t.Errorf("Expected to not find a selected key")
	}
	m, cmd := m.HandleMessage(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("enter"),
	})
	if cmd != nil {
		t.Errorf("Expected no commands")
	}
	m.Data = mock.Keys
	m, _ = m.HandleMessage(app.AccountSelected{Address: "ABC", Participation: &api.AccountParticipation{
		SelectionParticipationKey: nil,
		StateProofKey:             nil,
		VoteFirstValid:            0,
		VoteKeyDilution:           0,
		VoteLastValid:             0,
		VoteParticipationKey:      mock.VoteKey,
	}})
	d, active = m.SelectedKey()
	if !active {
		t.Errorf("Expected to find a selected key")
	}
	if d.Address != "ABC" {
		t.Errorf("Expected Address to be ABC, got %s", d.Address)
	}

	if m.Address != "ABC" {
		t.Errorf("Expected Address to be ABC, got %s", m.Address)
	}
}

func Test_Snapshot(t *testing.T) {
	t.Run("Visible", func(t *testing.T) {
		model := New("ABC", mock.Keys)
		model, _ = model.HandleMessage(tea.WindowSizeMsg{Width: 80, Height: 40})
		got := ansi.Strip(model.View())
		golden.RequireEqual(t, []byte(got))
	})
}

func Test_Messages(t *testing.T) {

	// Create the Model
	m := New("ABC", mock.Keys)
	//m, _ = m.Address = "ABC"
	tm := teatest.NewTestModel(
		t, m,
		teatest.WithInitialTermSize(80, 40),
	)

	// Wait for prompt to exit
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte("ABC"))
		},
		teatest.WithCheckInterval(time.Millisecond*100),
		teatest.WithDuration(time.Second*3),
	)

	// Emit a state message
	tm.Send(*test.GetState(nil))

	// Send delete finished
	tm.Send(app.DeleteFinished{
		Id: "1234",
	})

	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("enter"),
	})
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("esc"),
	})
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("ctrl+c"),
	})

	tm.Send(tea.QuitMsg{})

	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
