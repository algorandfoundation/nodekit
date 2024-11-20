package ui

import (
	"bytes"
	test2 "github.com/algorandfoundation/hack-tui/test"
	"github.com/algorandfoundation/hack-tui/ui/app"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
)

func Test_ViewportViewRender(t *testing.T) {
	client := test2.GetClient(false)
	state := test2.GetState(client)
	// Create the Model
	m, err := NewViewportViewModel(state, client)
	if err != nil {
		t.Fatal(err)
	}

	tm := teatest.NewTestModel(
		t, m,
		teatest.WithInitialTermSize(160, 40),
	)

	// Wait for prompt to exit
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte("Protocol Voting"))
		},
		teatest.WithCheckInterval(time.Millisecond*100),
		teatest.WithDuration(time.Second*3),
	)
	tm.Send(app.AccountSelected(
		state.Accounts["ABC"]))
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("left"),
	})

	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("right"),
	})

	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("right"),
	})

	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("left"),
	})

	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("left"),
	})
	// Send quit key
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("ctrl+c"),
	})

	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
