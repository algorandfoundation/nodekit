package ui

import (
	"github.com/algorandfoundation/hack-tui/ui/controls"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type ErrorViewModel struct {
	Height   int
	Width    int
	controls controls.Model
	Message  string
}

func NewErrorViewModel(message string) ErrorViewModel {
	return ErrorViewModel{
		Height:   0,
		Width:    0,
		controls: controls.New(" Error "),
	}
}

func (m ErrorViewModel) Init() tea.Cmd {
	return nil
}

func (m ErrorViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.HandleMessage(msg)
}

func (m ErrorViewModel) HandleMessage(msg tea.Msg) (ErrorViewModel, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height - 2
	}
	m.controls, cmd = m.controls.HandleMessage(msg)
	return m, cmd
}

func (m ErrorViewModel) View() string {
	pad := strings.Repeat("\n", max(0, m.Height/2-1))
	return lipgloss.JoinVertical(lipgloss.Center, pad, red.Render(m.Message), pad, m.controls.View())
}
