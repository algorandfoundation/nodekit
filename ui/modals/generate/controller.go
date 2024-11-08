package generate

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

func (m ViewModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.HandleMessage(msg)
}

func (m ViewModel) HandleMessage(msg tea.Msg) (*ViewModel, tea.Cmd) {
	var (
		cmd tea.Cmd
	)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return &m, EmitCancel(Cancel{})
		case "enter":
			return m.GenerateCmd()
		}

	}
	// Handle character input and blinking
	var val textinput.Model
	val, cmd = m.Input.Update(msg)

	m.Input = &val
	return &m, cmd
}
