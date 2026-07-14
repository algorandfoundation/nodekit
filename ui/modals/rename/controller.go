package rename

import (
	"strings"

	"github.com/algorandfoundation/nodekit/internal/algod/utils"
	"github.com/algorandfoundation/nodekit/ui/app"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

// Init initializes the ViewModel, starting the text input cursor blink.
func (m ViewModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update processes incoming messages and returns the updated model and command.
func (m ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.HandleMessage(msg)
}

// HandleMessage processes incoming messages, updates the ViewModel state, and
// returns an updated model and command.
func (m ViewModel) HandleMessage(msg tea.Msg) (ViewModel, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, app.EmitCloseOverlay()
		case "enter":
			value := strings.TrimSpace(m.Input.Value())
			if err := utils.SetNickname(m.Address, value); err != nil {
				m.InputError = "Error: " + err.Error()
				return m, nil
			}
			// Reflect the change in the shared state immediately so the
			// Accounts page renders the new nickname.
			if m.State != nil {
				if m.State.Nicknames == nil {
					m.State.Nicknames = map[string]string{}
				}
				if value == "" {
					delete(m.State.Nicknames, m.Address)
				} else {
					m.State.Nicknames[m.Address] = value
				}
			}
			m.InputError = ""
			return m, app.EmitCloseOverlay()
		}
	}

	m.Input, cmd = m.Input.Update(msg)
	return m, cmd
}
