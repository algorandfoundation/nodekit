package controls

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Init has no I/O right now
func (m Model) Init() tea.Cmd {
	return nil
}

// Update processes incoming messages, modifies the model state, and returns the updated model and command to execute.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.HandleMessage(msg)
}

// HandleMessage processes incoming messages, updates the model's state, and returns the updated model and a command to execute.
func (m Model) HandleMessage(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		if msg.Width != 0 && msg.Height != 0 {
			m.Width = msg.Width
			m.Height = lipgloss.Height(m.View())
		}
	}
	return m, nil
}
