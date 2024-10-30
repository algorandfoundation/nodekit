package ui

import (
	"github.com/algorandfoundation/hack-tui/ui/controls"
	"github.com/algorandfoundation/hack-tui/ui/style"
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
		Height:  0,
		Width:   0,
		Message: message,
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
		borderRender := style.Border.Render("")
		m.Width = max(0, msg.Width-lipgloss.Width(borderRender))
		m.Height = max(0, msg.Height-lipgloss.Height(borderRender))
	}

	return m, cmd
}

func (m ErrorViewModel) View() string {
	msgHeight := lipgloss.Height(m.Message)
	msgWidth := lipgloss.Width(m.Message)

	if msgWidth > m.Width/2 {
		m.Message = m.Message[0:m.Width/2] + "..."
		msgWidth = m.Width/2 + 3
	}

	msg := style.Red.Render(m.Message)
	padT := strings.Repeat("\n", max(0, (m.Height/2)-msgHeight))
	padL := strings.Repeat(" ", max(0, (m.Width-msgWidth)/2))

	text := lipgloss.JoinHorizontal(lipgloss.Left, padL, msg)
	render := style.ApplyBorder(m.Width, m.Height, "8").Render(lipgloss.JoinVertical(lipgloss.Center, padT, text))
	return style.WithNavigation(
		"( Waiting for recovery... )",
		style.WithTitle(
			"System Error",
			render,
		),
	)
}
