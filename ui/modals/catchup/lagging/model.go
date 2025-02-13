package lagging

import (
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/ui/app"
	"github.com/algorandfoundation/nodekit/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ViewModel struct {
	Height int
	Width  int
	// State is a pointer to an algod.StateModel, representing the state of the application including its configurations.
	State *algod.StateModel
}

func New(state *algod.StateModel) ViewModel {
	return ViewModel{
		State:  state,
		Height: 0,
		Width:  0,
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
	case app.FastCatchupStarted:
		return m, app.EmitCloseOverlay()
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			// Handle "yes" option
			return m, app.StartFastCatchupCmd(m.State)
		case "n", "N":
			return m, app.EmitCloseOverlay()
		}
	case tea.WindowSizeMsg:
		borderRender := style.Border.Render("")
		m.Width = max(0, msg.Width-lipgloss.Width(borderRender))
		m.Height = max(0, msg.Height-lipgloss.Height(borderRender))
	}

	return m, cmd
}

// Title returns the static title string "Delete Key" for the ViewModel.
func (m ViewModel) Title() string {
	return "( Out of Sync )"
}

// BorderColor returns the border color as a string, typically used for rendering styled components in the ViewModel.
func (m ViewModel) BorderColor() string {
	return "9"
}

// Controls returns a formatted string displaying the available control options (yes or no) with styled color representations.
func (m ViewModel) Controls() string {
	return "( " + style.Green.Render("(y)es") + " | " + style.Red.Render("(n)o") + " )"
}

// Body returns the formatted body content of the ViewModel, including participation key details or a default message.
func (m ViewModel) Body() string {
	return lipgloss.NewStyle().Padding(1).Render(lipgloss.JoinVertical(lipgloss.Center,
		"Your node is significantly behind the network.\n Would you like to perform a fast-catchup?\n",
	))

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
			style.ApplyBorder(width+2, height-4, m.BorderColor()).
				PaddingRight(1).
				PaddingLeft(1).
				Render(m.Body()),
		),
	)

}
