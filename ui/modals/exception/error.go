package exception

import (
	"github.com/algorandfoundation/nodekit/ui/app"
	"github.com/algorandfoundation/nodekit/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type ViewModel struct {
	Height  int
	Width   int
	Message string
}

func New(message string) ViewModel {
	return ViewModel{
		Height:  0,
		Width:   0,
		Message: message,
	}
}

func (m ViewModel) Init() tea.Cmd {
	return nil
}

func (m ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.HandleMessage(msg)
}

func (m ViewModel) HandleMessage(msg tea.Msg) (ViewModel, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	// Handle errors make ensure the modal is visible
	case error:
		m.Message = msg.Error()
		return m, app.EmitShowModal(app.ExceptionModal)
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, app.EmitCloseOverlay()

		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}

	return m, cmd
}

func (m ViewModel) Title() string {
	return "Error"
}
func (m ViewModel) BorderColor() string {
	return "1"
}
func (m ViewModel) Controls() string {
	return "( esc )"
}
func (m ViewModel) Body() string {
	return ansi.Hardwrap(style.Red.Render(m.Message), m.Width, false)
}

// View renders the ViewModel as a styled string, incorporating title, controls, and body content with dynamic borders.
func (m ViewModel) View() string {
	body := m.Body()
	width := lipgloss.Width(body)
	height := lipgloss.Height(body)
	return style.WithNavigation(
		m.Controls(),
		style.WithTitle(
			m.Title(),
			// Apply the Borders with the Padding
			style.ApplyBorder(width+2, height+2, m.BorderColor()).
				Padding(1).
				Render(m.Body()),
		),
	)

}
