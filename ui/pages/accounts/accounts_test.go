package accounts

import (
	"bytes"
	"github.com/algorandfoundation/hack-tui/api"
	"github.com/algorandfoundation/hack-tui/internal"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/charmbracelet/x/exp/teatest"
	"testing"
	"time"
)

func Test_Snapshot(t *testing.T) {
	t.Run("Visible", func(t *testing.T) {
		model := New(&internal.StateModel{
			Status:            internal.StatusModel{},
			Metrics:           internal.MetricsModel{},
			Accounts:          nil,
			ParticipationKeys: nil,
			Admin:             false,
			Watching:          false,
		})
		got := ansi.Strip(model.View())
		golden.RequireEqual(t, []byte(got))
	})
}

func Test_Messages(t *testing.T) {
	var testKeys = []api.ParticipationKey{
		{
			Address:             "ABC",
			EffectiveFirstValid: nil,
			EffectiveLastValid:  nil,
			Id:                  "",
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
	m := New(sm)

	tm := teatest.NewTestModel(
		t, m,
		teatest.WithInitialTermSize(80, 40),
	)

	// Wait for prompt to exit
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte("(k)eys"))
		},
		teatest.WithCheckInterval(time.Millisecond*100),
		teatest.WithDuration(time.Second*3),
	)

	tm.Send(*sm)

	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("enter"),
	})

	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("ctrl+c"),
	})

	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
