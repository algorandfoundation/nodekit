package catchup

import (
	"fmt"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/ui/app"
	"github.com/algorandfoundation/nodekit/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"time"
)

type ViewModel struct {
	State  *algod.StateModel
	Height int
	Width  int
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
	case app.FastCatchupStopped:
		return m, app.EmitCloseOverlay()
	case tea.KeyMsg:
		switch msg.String() {
		// TODO: Maybe abort?
		case "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}

	return m, cmd
}

// Title returns the static title string "Delete Key" for the ViewModel.
func (m ViewModel) Title() string {
	return "Fast Catchup"
}

// BorderColor returns the border color as a string, typically used for rendering styled components in the ViewModel.
func (m ViewModel) BorderColor() string {
	return "7"
}

// Controls returns a formatted string displaying the available control options (yes or no) with styled color representations.
func (m ViewModel) Controls() string {
	return ""
}

// Body returns the formatted body content of the ViewModel, including participation key details or a default message.
func (m ViewModel) Body() string {
	return style.LightBlue(lipgloss.JoinVertical(lipgloss.Top,
		"Please wait while your node syncs with the network.",
		"This process can take up to an hour.",
		"",
		fmt.Sprintf("Accounts Processed:   %d / %d", m.State.Status.CatchpointAccountsProcessed, m.State.Status.CatchpointAccountsTotal),
		fmt.Sprintf("Accounts Verified:    %d / %d", m.State.Status.CatchpointAccountsVerified, m.State.Status.CatchpointAccountsTotal),
		fmt.Sprintf("Key Values Processed: %d / %d", m.State.Status.CatchpointKeyValueProcessed, m.State.Status.CatchpointKeyValueTotal),
		fmt.Sprintf("Key Values Verified:  %d / %d", m.State.Status.CatchpointKeyValueVerified, m.State.Status.CatchpointKeyValueTotal),
		fmt.Sprintf("Downloaded blocks:    %d / %d", m.State.Status.CatchpointBlocksAcquired, m.State.Status.CatchpointBlocksTotal),
		"",
		fmt.Sprintf("Sync Time: %ds", m.State.Status.SyncTime/int(time.Second)),
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
