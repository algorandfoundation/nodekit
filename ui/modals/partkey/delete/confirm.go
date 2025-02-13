package delete

import (
	"github.com/algorandfoundation/nodekit/api"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/ui/app"
	"github.com/algorandfoundation/nodekit/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Init initializes the ViewModel and returns a tea.Cmd to start the program or execute initial commands.
func (m ViewModel) Init() tea.Cmd {
	return nil
}

// ViewModel represents the main structure containing state, dimensions, and participation data for view rendering.
type ViewModel struct {
	// Width defines the horizontal dimension of the ViewModel, typically measured in units such as characters or pixels.
	Width int
	// Height defines the vertical dimension of the ViewModel, commonly measured in units such as characters or pixels.
	Height int
	// Participation is a pointer to an api.ParticipationKey representing a participation key used by the node.
	Participation *api.ParticipationKey
	// State is a pointer to an algod.StateModel, representing the state of the application including its configurations.
	State *algod.StateModel
}

// New initializes and returns a new ViewModel with specified state and participation key, setting dimensions to default values.
func New(state *algod.StateModel, participation *api.ParticipationKey) ViewModel {
	return ViewModel{
		Width:         0,
		Height:        0,
		Participation: participation,
		State:         state,
	}
}

// Update processes a message, updates the ViewModel state, and returns the updated model along with a possible command.
func (m ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.HandleMessage(msg)
}

// HandleMessage processes incoming messages, updates ViewModel state, and returns the updated model alongside a command.
func (m ViewModel) HandleMessage(msg tea.Msg) (ViewModel, tea.Cmd) {
	switch msg := msg.(type) {
	// Handle Confirmation Dialog Delete Finished
	case app.DeleteFinished:
		return m, app.EmitCloseOverlay()
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "n":
			return m, app.EmitCancelOverlay()
		case "y":
			// Emit the delete request
			return m, app.EmitDeleteKey(m.State.Context, m.State.Client, m.Participation.Id)
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}
	return m, nil
}

// Title returns the static title string "Delete Key" for the ViewModel.
func (m ViewModel) Title() string {
	return "Delete Key"
}

// BorderColor returns the border color as a string, typically used for rendering styled components in the ViewModel.
func (m ViewModel) BorderColor() string {
	return "9"
}

// Controls returns a string representation of the available control options for the ViewModel.
func (m ViewModel) Controls() string {
	return "| (esc) |"
}

// Navigation returns a formatted string displaying the available control options (yes or no) with styled color representations.
func (m ViewModel) Navigation() string {
	return "( " + style.Green.Render("(y)es") + " | " + style.Red.Render("(n)o") + " )"
}

// Body returns the formatted body content of the ViewModel, including participation key details or a default message.
func (m ViewModel) Body() string {
	if m.Participation == nil {
		return "No key selected"
	}
	return lipgloss.NewStyle().Padding(1).Render(lipgloss.JoinVertical(lipgloss.Center,
		"Are you sure you want to delete this key from your node?\n",
		style.Cyan.Render("Account Address:"),
		m.Participation.Address+"\n",
		style.Cyan.Render("Participation Key:"),
		m.Participation.Id,
	))

}

// View renders the ViewModel as a styled string, incorporating title, controls, and body content with dynamic borders.
func (m ViewModel) View() string {
	body := m.Body()
	width := lipgloss.Width(body)
	height := lipgloss.Height(body)
	return style.WithControls(
		m.Controls(),
		style.WithNavigation(
			m.Navigation(),
			style.WithTitle(
				m.Title(),
				// Apply the Borders with the Padding
				style.ApplyBorder(width+2, height-4, m.BorderColor()).
					PaddingRight(1).
					PaddingLeft(1).
					Render(m.Body()),
			),
		))

}
