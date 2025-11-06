package hybrid

import (
	"os"

	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/internal/algod/utils"
	"github.com/algorandfoundation/nodekit/ui/app"
	"github.com/algorandfoundation/nodekit/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
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
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "enter", "p":
			return m, app.EmitCloseOverlay()
		case "d":
			err := utils.DontShowHybridPopUp()
			if err != nil {
				log.Warnf("unable to disable hybrid popup: %s", err)
			}
			return m, app.EmitCloseOverlay()
		case "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
	}

	return m, cmd
}

// Title returns the static title string "NodeKit Information" for the ViewModel.
func (m ViewModel) Title() string {
	return "NodeKit Information"
}

// BorderColor returns the border color as a string, typically used for rendering styled components in the ViewModel.
func (m ViewModel) BorderColor() string {
	return "7"
}

// Controls returns a formatted string displaying the available control options (yes or no) with styled color representations.
func (m ViewModel) Controls() string {
	hybridEnabled := m.State.Config.EnableP2PHybridMode != nil && *m.State.Config.EnableP2PHybridMode
	controls := "| "
	if !hybridEnabled && utils.ShowHybridPopUp() {
		controls += style.Red.Render("(d)on't show again") + " | "
	}
	controls += style.Red.Render("(enter) to close") + " |"
	return controls
}

// Body returns the formatted body content of the ViewModel, including information about P2P Hybrid being available with a URL to learn more, and how to enable it.
func (m ViewModel) Body() string {
	// TODO: Update this link to the actual NodeKit P2P Hybrid Mode documentation before release.
	link := "https://d.nodekit.run/abcdef"
	return lipgloss.JoinVertical(lipgloss.Center,
		"",
		"Did you know P2P Hybrid Mode is now available in NodeKit?",
		"",
		"Read more by visiting:",
		style.LightBlue(style.WithHyperlink(link, link)),
		"",
		"Or by running:",
		style.LightBlue(os.Args[0]+" configure algod -h"),
		"",
		"To Enable P2P Hybrid Mode:",
		style.LightBlue(os.Args[0]+" configure algod --hybrid=true"),
		"",
	)
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
