package ui

import (
	"bytes"
	"github.com/algorandfoundation/nodekit/internal/test"
	"github.com/algorandfoundation/nodekit/ui/app"
	uitest "github.com/algorandfoundation/nodekit/ui/internal/test"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
)

func Test_ViewportViewRender(t *testing.T) {
	client := test.GetClient(false)
	state := uitest.GetState(client)
	// Create the Model
	m, err := NewViewportViewModel(state)
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
			return bytes.Contains(bts, []byte("Protocol Upgrade"))
		},
		teatest.WithCheckInterval(time.Millisecond*100),
		teatest.WithDuration(time.Second*3),
	)
	acc := state.Accounts["ABC"]
	tm.Send(app.AccountSelected(
		&acc))
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
		Runes: []rune("q"),
	})

	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
