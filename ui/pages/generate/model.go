package generate

import (
	"github.com/algorandfoundation/hack-tui/api"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type ViewModel struct {
	Address    string
	Inputs     []textinput.Model
	client     *api.ClientWithResponses
	focusIndex int
	cursorMode cursor.Mode
}

func (m *ViewModel) updateInputs(msg tea.Msg) tea.Cmd {
	cmds := make([]tea.Cmd, len(m.Inputs))

	// Only text inputs with Focus() set will respond, so it's safe to simply
	// update all of them here without any further logic.
	for i := range m.Inputs {
		m.Inputs[i], cmds[i] = m.Inputs[i].Update(msg)
	}

	return tea.Batch(cmds...)
}

func New(address string, client *api.ClientWithResponses) ViewModel {
	m := ViewModel{
		Address: address,
		Inputs:  make([]textinput.Model, 3),
		client:  client,
	}

	var t textinput.Model
	for i := range m.Inputs {
		t = textinput.New()
		t.Cursor.Style = cursorStyle
		t.CharLimit = 68

		switch i {
		case 0:
			t.Placeholder = "Wallet Address or NFD"
			t.Focus()
			t.PromptStyle = focusedStyle
			t.TextStyle = focusedStyle
			t.CharLimit = 68
		case 1:
			t.Placeholder = "First Valid Round"
		case 2:
			t.Placeholder = "Last"
		}

		m.Inputs[i] = t
	}

	return m
}
