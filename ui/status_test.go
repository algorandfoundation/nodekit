package ui

import (
	"bytes"
	"github.com/algorandfoundation/hack-tui/api"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"strings"
	"testing"
	"time"
)

func Test_ExecuteInvalidStatusCommand(t *testing.T) {
	client, err := api.NewClientWithResponses("http://255.255.255.255:4001")
	if err != nil {
		t.Fatal(err)
	}

	// Test Invalid Node
	_, err = MakeStatusViewModel(client)
	if !strings.Contains(err.Error(), "dial tcp 255.255.255.255:4001") {
		t.Fatal(err)
	}
}
func Test_ExecuteStatusCommand(t *testing.T) {
	client, err := api.NewClientWithResponses("https://mainnet-api.4160.nodely.dev")
	if err != nil {
		t.Fatal(err)
	}

	var m tea.Model
	m, err = MakeStatusViewModel(client)

	if err != nil {
		t.Fatal(err)
	}

	tm := teatest.NewTestModel(
		t, m,
		teatest.WithInitialTermSize(300, 100),
	)

	// Wait for prompt to exit
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte("Press q to quit."))
		},
		teatest.WithCheckInterval(time.Millisecond*100),
		teatest.WithDuration(time.Second*3),
	)

	// Send a block update
	tm.Send(tea.Msg(uint64(123)))

	// Send hide key
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("h"),
	})

	// Send quit key
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("q"),
	})

	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
