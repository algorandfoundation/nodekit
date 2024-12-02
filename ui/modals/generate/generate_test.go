package generate

import (
	"bytes"
	"github.com/algorandfoundation/hack-tui/ui/internal/test"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/exp/golden"
	"github.com/charmbracelet/x/exp/teatest"
	"testing"
	"time"
)

func Test_New(t *testing.T) {
	m := New("ABC", test.GetState(nil))

	m.SetAddress("TUIDKH2C7MUHZDD77MAMUREJRKNK25SYXB7OAFA6JFBB24PEL5UX4S4GUU")

	if m.Address != "TUIDKH2C7MUHZDD77MAMUREJRKNK25SYXB7OAFA6JFBB24PEL5UX4S4GUU" {
		t.Error("Did not set address")
	}

	m.SetStep(AddressStep)
	if m.Step != AddressStep {
		t.Error("Did not advance to address step")
	}
	if m.Controls != "( esc to cancel )" {
		t.Error("Did not set controls")
	}

	m.SetStep(DurationStep)
	m.InputTwo.SetValue("1")

	m, cmd := m.HandleMessage(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("enter"),
	})
	if cmd == nil {
		t.Error("Did not return the generate command")
	}
	if m.Step != WaitingStep {
		t.Error("Did not advance to waiting step")
	}

	m.SetStep(DurationStep)
	m.Range = Week
	m.InputTwo.SetValue("1")
	m, cmd = m.HandleMessage(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("enter"),
	})
	if cmd == nil {
		t.Error("Did not return the generate command")
	}

	m.SetStep(DurationStep)
	m.Range = Month
	m.InputTwo.SetValue("1")
	m, cmd = m.HandleMessage(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("enter"),
	})
	if cmd == nil {
		t.Error("Did not return the generate command")
	}

	m.SetStep(DurationStep)
	m.Range = Year
	m.InputTwo.SetValue("1")
	m, cmd = m.HandleMessage(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("enter"),
	})
	if cmd == nil {
		t.Error("Did not return the generate command")
	}
}

func Test_Snapshot(t *testing.T) {
	t.Run("Visible", func(t *testing.T) {
		model := New("ABC", test.GetState(nil))
		got := ansi.Strip(model.View())
		golden.RequireEqual(t, []byte(got))
	})
	t.Run("Duration", func(t *testing.T) {
		model := New("ABC", test.GetState(nil))
		model.SetStep(DurationStep)
		got := ansi.Strip(model.View())
		golden.RequireEqual(t, []byte(got))
	})
	t.Run("Waiting", func(t *testing.T) {
		model := New("ABC", test.GetState(nil))
		model.SetStep(WaitingStep)
		got := ansi.Strip(model.View())
		golden.RequireEqual(t, []byte(got))
	})
}

func Test_Messages(t *testing.T) {
	// Create the Model
	m := New("ABC", test.GetState(nil))
	tm := teatest.NewTestModel(
		t, m,
		teatest.WithInitialTermSize(80, 40),
	)

	// Wait for prompt to exit
	teatest.WaitFor(
		t, tm.Output(),
		func(bts []byte) bool {
			return bytes.Contains(bts, []byte("Create keys required to participate in Algorand consensus."))
		},
		teatest.WithCheckInterval(time.Millisecond*100),
		teatest.WithDuration(time.Second*3),
	)

	// Enter into duration mode
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("enter"),
	})

	// Rotate the durations
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("s"),
	})
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("s"),
	})
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("s"),
	})
	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("s"),
	})

	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("1"),
	})

	tm.Send(tea.KeyMsg{
		Type:  tea.KeyRunes,
		Runes: []rune("esc"),
	})
	tm.Send(tea.QuitMsg{})

	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second))
}
