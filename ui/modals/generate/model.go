package generate

import (
	"github.com/algorandfoundation/hack-tui/internal"
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
)

type ViewModel struct {
	Width  int
	Height int

	Address string
	Input   *textinput.Model

	Title       string
	Controls    string
	BorderColor string

	State      *internal.StateModel
	cursorMode cursor.Mode
}

func (m ViewModel) SetAddress(address string) {
	m.Address = address
	m.Input.SetValue(address)
}

func New(address string, state *internal.StateModel) *ViewModel {
	input := textinput.New()
	m := ViewModel{
		Address:     address,
		State:       state,
		Input:       &input,
		Title:       "Generate Participation Key",
		Controls:    "( esc to cancel )",
		BorderColor: "2",
	}
	input.Cursor.Style = cursorStyle
	input.CharLimit = 68
	input.Placeholder = "Wallet Address"
	input.Focus()
	input.PromptStyle = focusedStyle
	input.TextStyle = focusedStyle
	return &m
}
