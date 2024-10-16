package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

// ControlViewModel handles the
type ControlViewModel struct {
	IsVisible bool
	ViewWidth int
}

// Init has no I/O right now
func (m ControlViewModel) Init() tea.Cmd {
	return nil
}

func (m ControlViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.HandleMessage(msg)
}

func (m ControlViewModel) HandleMessage(msg tea.Msg) (ControlViewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ViewWidth = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		// Hide on h keypress
		case "h":
			m.IsVisible = !m.IsVisible
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

// View renders the string
func (m ControlViewModel) View() string {
	if !m.IsVisible {
		return ""
	}
	info := infoStyle.Render(" (q)uit | (d)elete | (g)enerate | (t)xn | (h)ide ")
	difference := m.ViewWidth - lipgloss.Width(info)

	line := strings.Repeat("─", max(0, difference/2))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info, line)
}

func MakeControlViewModel() ControlViewModel {
	return ControlViewModel{
		IsVisible: true,
		ViewWidth: 80,
	}
}
