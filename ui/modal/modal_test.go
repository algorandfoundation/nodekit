package modal

import (
	"bytes"
	"errors"
	"github.com/algorandfoundation/hack-tui/internal/test/mock"
	"github.com/algorandfoundation/hack-tui/ui/app"
	"github.com/algorandfoundation/hack-tui/ui/internal/test"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/charmbracelet/x/exp/teatest"
	"testing"
	"time"
)

func Test_Snapshot(t *testing.T) {
	t.Run("NoKey", func(t *testing.T) {
		model := New(lipgloss.NewStyle().Width(80).Height(80).Render(""), true, test.GetState(nil))

		model, _ = model.HandleMessage(tea.WindowSizeMsg{Width: 80, Height: 40})
		got := ansi.Strip(model.View())
		golden.RequireEqual(t, []byte(got))
	})
	t.Run("InfoModal", func(t *testing.T) {
		model := New(lipgloss.NewStyle().Width(80).Height(80).Render(""), true, test.GetState(nil))
		model.SetKey(&mock.Keys[0])
		model.SetType(app.InfoModal)
		model, _ = model.HandleMessage(tea.WindowSizeMsg{Width: 80, Height: 40})
		got := ansi.Strip(model.View())
		golden.RequireEqual(t, []byte(got))
	})
	t.Run("ConfirmModal", func(t *testing.T) {
		model := New(lipgloss.NewStyle().Width(80).Height(80).Render(""), true, test.GetState(nil))
		model.SetKey(&mock.Keys[0])
		model.SetType(app.ConfirmModal)
		model, _ = model.HandleMessage(tea.WindowSizeMsg{Width: 80, Height: 40})
		got := ansi.Strip(model.View())
		golden.RequireEqual(t, []byte(got))
	})
	t.Run("ExceptionModal", func(t *testing.T) {
		model := New(lipgloss.NewStyle().Width(80).Height(80).Render(""), true, test.GetState(nil))
		model.SetKey(&mock.Keys[0])
		model.SetType(app.ExceptionModal)
		model, _ = model.HandleMessage(errors.New("test error"))
		got := ansi.Strip(model.View())
		golden.RequireEqual(t, []byte(got))
	})
	t.Run("GenerateModal", func(t *testing.T) {
		model := New(lipgloss.NewStyle().Width(80).Height(80).Render(""), true, test.GetState(nil))
		model.SetKey(&mock.Keys[0])
		model.SetAddress("ABC")
		model.SetType(app.GenerateModal)
		model, _ = model.HandleMessage(tea.WindowSizeMsg{Width: 80, Height: 40})
		got := ansi.Strip(model.View())
		golden.RequireEqual(t, []byte(got))
	})

	t.Run("TransactionModal", func(t *testing.T) {
		model := New(lipgloss.NewStyle().Width(80).Height(80).Render(""), true, test.GetState(nil))
		model.State.Status.Network = "testnet-v1.0"
		model.SetKey(&mock.Keys[0])
		model.SetActive(true)
		model.SetType(app.TransactionModal)
		model, _ = model.HandleMessage(tea.WindowSizeMsg{Width: 80, Height: 40})
		got := ansi.Strip(model.View())
		golden.RequireEqual(t, []byte(got))
	})
}

func Test_Messages(t *testing.T) {
	model := New(lipgloss.NewStyle().Width(80).Height(80).Render(""), true, test.GetState(nil))
	model.SetKey(&mock.Keys[0])
	model.SetAddress("ABC")
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

	tm.Send(app.ModalEvent{
		Key:     nil,
		Active:  false,
		Address: "ABC",
		Err:     nil,
		Type:    app.InfoModal,
	})
	tm.Send(app.ModalEvent{
		Key:     nil,
		Active:  false,
		Address: "ABC",
		Err:     nil,
		Type:    app.CancelModal,
	})
	tm.Send(app.ModalEvent{
		Key:     nil,
		Active:  false,
		Address: "ABC",
		Err:     nil,
		Type:    app.GenerateModal,
	})
	tm.Send(app.ModalEvent{
		Key:     nil,
		Active:  false,
		Address: "ABC",
		Err:     nil,
		Type:    app.CancelModal,
	})
	tm.Send(app.ModalEvent{
		Key:     nil,
		Active:  false,
		Address: "ABC",
		Err:     nil,
		Type:    app.ConfirmModal,
	})
	tm.Send(app.ModalEvent{
		Key:     nil,
		Active:  false,
		Address: "ABC",
		Err:     nil,
		Type:    app.CancelModal,
	})
	tm.Send(app.ModalEvent{
		Key:     nil,
		Active:  false,
		Address: "ABC",
		Err:     nil,
		Type:    app.TransactionModal,
	})
	tm.Send(app.ModalEvent{
		Key:     nil,
		Active:  false,
		Address: "ABC",
		Err:     nil,
		Type:    app.CancelModal,
	})

	tm.Send(app.ModalEvent{
		Key:     nil,
		Active:  false,
		Address: "ABC",
		Err:     nil,
		Type:    app.CloseModal,
	})
	tm.Send(tea.QuitMsg{})

	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
