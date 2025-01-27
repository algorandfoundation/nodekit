package modal

import (
	"bytes"
	"errors"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/internal/test/mock"
	"github.com/algorandfoundation/nodekit/ui/app"
	"github.com/algorandfoundation/nodekit/ui/internal/test"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/exp/teatest"
	"testing"
	"time"
)

func Test_Snapshot(t *testing.T) {
	t.Skip("TODO:")
}

func Test_Messages(t *testing.T) {
	model := New(lipgloss.NewStyle().Width(80).Height(80).Render(""), true, test.GetState(nil))
	model.SetKey(&mock.Keys[0])
	//model.SetAddress("ABC")
	model.SetType(app.InfoModal)
	tm := teatest.NewTestModel(
		t, model,
		teatest.WithInitialTermSize(80, 40),
	)

	// Wait for prompt to exit
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte("State Proof Key: VEVTVEtFWQ"))
		},
		teatest.WithCheckInterval(time.Millisecond*100),
		teatest.WithDuration(time.Second*3),
	)

	tm.Send(errors.New("Something else went wrong"))

	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("d"),
	})

	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("esc"),
	})

	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("o"),
	})

	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("esc"),
	})

	tm.Send(app.InfoModal)

	tm.Send(app.DeleteFinished{
		Err: nil,
		Id:  mock.Keys[0].Id,
	})

	delError := errors.New("Something went wrong")
	tm.Send(app.DeleteFinished{
		Err: &delError,
		Id:  "",
	})

	tm.Send(app.KeySelectedEvent{
		Key:    nil,
		Active: false,
	})
	tm.Send(app.AccountSelected(&algod.Account{
		Address: "ABC",
	}))
	tm.Send(app.KeySelectedEvent{
		Key:    nil,
		Active: false,
	})
	tm.Send(tea.QuitMsg{})

	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
