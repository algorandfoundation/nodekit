package rename

import (
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/charmbracelet/bubbles/textinput"
)

// ViewModel holds the state for the modal used to assign a local nickname to an
// account. Nicknames are stored locally in the NodeKit settings file and are a
// display convenience only.
type ViewModel struct {
	Width  int
	Height int

	// Address is the account currently being named.
	Address string

	// Input collects the nickname text.
	Input textinput.Model
	// InputError holds a validation/persistence error to surface to the user.
	InputError string

	// State is the shared application state, updated in place when a nickname
	// is set so the Accounts page reflects the change.
	State *algod.StateModel
}

// New creates a rename ViewModel bound to the provided application state.
func New(state *algod.StateModel) ViewModel {
	m := ViewModel{
		State: state,
		Input: textinput.New(),
	}
	m.Input.Cursor.Style = cursorStyle
	m.Input.CharLimit = 32
	m.Input.Placeholder = "Nickname"
	m.Input.PromptStyle = focusedStyle
	m.Input.TextStyle = focusedStyle
	return m
}

// SetAddress targets the modal at an account address, prefilling the input with
// any existing nickname and focusing it.
func (m *ViewModel) SetAddress(address string) {
	m.Address = address
	m.InputError = ""
	current := ""
	if m.State != nil {
		current = m.State.Nicknames[address]
	}
	m.Input.SetValue(current)
	m.Input.CursorEnd()
	m.Input.Focus()
}
