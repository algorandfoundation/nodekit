package keys

import (
	"bytes"
	"github.com/algorandfoundation/hack-tui/api"
	"github.com/algorandfoundation/hack-tui/internal"
	"github.com/algorandfoundation/hack-tui/ui/app"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/charmbracelet/x/exp/teatest"
	"testing"
	"time"
)

var testVoteKey = []byte("TESTKEY")
var testKeys = []api.ParticipationKey{
	{
		Address:             "ABC",
		EffectiveFirstValid: nil,
		EffectiveLastValid:  nil,
		Id:                  "123",
		Key: api.AccountParticipation{
			SelectionParticipationKey: nil,
			StateProofKey:             nil,
			VoteFirstValid:            0,
			VoteKeyDilution:           0,
			VoteLastValid:             0,
			VoteParticipationKey:      testVoteKey,
		},
		LastBlockProposal: nil,
		LastStateProof:    nil,
		LastVote:          nil,
	},
	{
		Address:             "ABC",
		EffectiveFirstValid: nil,
		EffectiveLastValid:  nil,
		Id:                  "1234",
		Key: api.AccountParticipation{
			SelectionParticipationKey: nil,
			StateProofKey:             nil,
			VoteFirstValid:            0,
			VoteKeyDilution:           0,
			VoteLastValid:             0,
			VoteParticipationKey:      nil,
		},
		LastBlockProposal: nil,
		LastStateProof:    nil,
		LastVote:          nil,
	},
}

func Test_New(t *testing.T) {
	m := New("ABC", nil)
	if m.Address != "ABC" {
		t.Errorf("Expected Address to be ABC, got %s", m.Address)
	}
	d, active := m.SelectedKey()
	if active {
		t.Errorf("Expected to not find a selected key")
	}
	m, err := m.HandleMessage(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("enter"),
	})

	if err != nil {
		t.Errorf("Expected no error")
	}
	m.Data = &testKeys
	m, _ = m.HandleMessage(app.AccountSelected{Address: "ABC", Participation: &api.AccountParticipation{
		SelectionParticipationKey: nil,
		StateProofKey:             nil,
		VoteFirstValid:            0,
		VoteKeyDilution:           0,
		VoteLastValid:             0,
		VoteParticipationKey:      testVoteKey,
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
		model := New("ABC", &testKeys)
		model, _ = model.HandleMessage(tea.WindowSizeMsg{Width: 80, Height: 40})
		got := ansi.Strip(model.View())
		golden.RequireEqual(t, []byte(got))
	})
}

func Test_Messages(t *testing.T) {
	sm := &internal.StateModel{
		Status:            internal.StatusModel{},
		Metrics:           internal.MetricsModel{},
		Accounts:          nil,
		ParticipationKeys: &testKeys,
		Admin:             false,
		Watching:          false,
	}
	values := make(map[string]internal.Account)
	for _, key := range *sm.ParticipationKeys {
		val, ok := values[key.Address]
		if !ok {
			values[key.Address] = internal.Account{
				Address: key.Address,
				Status:  "Offline",
				Balance: 0,
				Expires: time.Unix(0, 0),
				Keys:    1,
			}
		} else {
			val.Keys++
			values[key.Address] = val
		}
	}
	sm.Accounts = values
	// Create the Model
	m := New("ABC", &testKeys)
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
	tm.Send(*sm)

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
