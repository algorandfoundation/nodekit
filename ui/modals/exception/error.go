package exception

import (
	"github.com/algorandfoundation/algorun-tui/ui/app"
	"github.com/algorandfoundation/algorun-tui/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

type ViewModel struct {
	Height  int
	Width   int
	Message string

	Title       string
	BorderColor string
	Controls    string
	Navigation  string
}

func New(message string) *ViewModel {
	return &ViewModel{
		Height:      0,
		Width:       0,
		Message:     message,
		Title:       "Error",
		BorderColor: "1",
		Controls:    "( esc )",
		Navigation:  "",
	}
}

func (m ViewModel) Init() tea.Cmd {
	return nil
}

func (m ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.HandleMessage(msg)
}

func (m ViewModel) HandleMessage(msg tea.Msg) (*ViewModel, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case error:
		m.Message = msg.Error()
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return &m, app.EmitModalEvent(app.ModalEvent{
				Type: app.CancelModal,
			})

		}
	case tea.WindowSizeMsg:
		borderRender := style.Border.Render("")
		m.Width = max(0, msg.Width-lipgloss.Width(borderRender))
		m.Height = max(0, msg.Height-lipgloss.Height(borderRender))
	}

	return &m, cmd
}

func (m ViewModel) View() string {
	return ansi.Hardwrap(style.Red.Render(m.Message), m.Width, false)
}
