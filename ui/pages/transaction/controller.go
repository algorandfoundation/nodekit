package transaction

import (
	"github.com/algorandfoundation/hack-tui/api"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m ViewModel) Init() tea.Cmd {
	return nil
}

func (m ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.HandleMessage(msg)
}

// HandleMessage is called by the viewport to update it's Model
func (m ViewModel) HandleMessage(msg tea.Msg) (ViewModel, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	// When the participation key updates, set the models data
	case api.ParticipationKey:
		m.Data = msg
	// Handle View Size changes
	case tea.WindowSizeMsg:
		if msg.Width != 0 && msg.Height != 0 {
			m.Width = msg.Width
			m.Height = max(0, msg.Height-lipgloss.Height(m.controls.View()))
		}
	}

	// Pass messages to controls
	m.controls, cmd = m.controls.HandleMessage(msg)
	return m, cmd
}
