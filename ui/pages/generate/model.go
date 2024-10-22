package generate

import (
	"github.com/algorandfoundation/hack-tui/api"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
)

type ViewModel struct {
	Address    string
	Inputs     []textinput.Model
	client     *api.ClientWithResponses
	focusIndex int
	cursorMode cursor.Mode
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
