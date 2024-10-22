package ui

import (
	"github.com/algorandfoundation/hack-tui/internal"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strconv"
	"strings"
)

// ProtocolViewModel includes the internal.StatusModel and internal.MetricsModel
type ProtocolViewModel struct {
	Data           internal.StatusModel
	TerminalWidth  int
	TerminalHeight int
	IsVisible      bool
}

// Init has no I/O right now
func (m ProtocolViewModel) Init() tea.Cmd {
	return nil
}

// Update applies a message to the model and returns an updated model and command.
func (m ProtocolViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.HandleMessage(msg)
}

// HandleMessage processes incoming messages and updates the ProtocolViewModel's state.
// It handles tea.WindowSizeMsg to update ViewWidth and tea.KeyMsg for key events like 'h' to toggle visibility and 'q' or 'ctrl+c' to quit.
func (m ProtocolViewModel) HandleMessage(msg tea.Msg) (ProtocolViewModel, tea.Cmd) {
	switch msg := msg.(type) {
	// Handle a Status Update
	case internal.StatusModel:
		m.Data = msg
		return m, nil
	// Update Viewport Size
	case tea.WindowSizeMsg:
		m.TerminalWidth = msg.Width
		m.TerminalHeight = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		// The H key should hide the render
		case "h":
			m.IsVisible = !m.IsVisible
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	// Return the updated model to the Bubble Tea runtime for processing.
	return m, nil
}

// View renders the view for the ProtocolViewModel according to the current state and dimensions.
func (m ProtocolViewModel) View() string {
	if !m.IsVisible {
		return ""
	}
	if m.TerminalWidth <= 0 {
		return "Loading...\n\n\n\n\n\n"
	}
	beginning := blue.Render(" Node: ") + m.Data.Version

	isCompact := m.TerminalWidth < 90

	if isCompact && m.TerminalHeight < 26 {
		return ""
	}

	end := ""
	if m.Data.NeedsUpdate && !isCompact {
		end += green.Render("[UPDATE AVAILABLE] ")
	}

	var size int
	if isCompact {
		size = m.TerminalWidth
	} else {
		size = m.TerminalWidth / 2
	}

	middle := strings.Repeat(" ", max(0, size-(lipgloss.Width(beginning)+lipgloss.Width(end)+2)))

	var rows []string
	// Last Round
	rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Left, beginning, middle, end))
	if !isCompact {
		rows = append(rows, "")
	}
	rows = append(rows, blue.Render(" Network: ")+m.Data.Network)
	if !isCompact {
		rows = append(rows, "")
	}
	rows = append(rows, blue.Render(" Protocol Voting: ")+strconv.FormatBool(m.Data.Voting))

	if isCompact && m.Data.NeedsUpdate {
		rows = append(rows, blue.Render(" Upgrade Available: ")+green.Render(strconv.FormatBool(m.Data.NeedsUpdate)))
	}
	return topSections(max(0, size)).Render(lipgloss.JoinVertical(lipgloss.Left,
		rows...,
	))
}

// MakeProtocolViewModel constructs a ProtocolViewModel using a given StatusModel and predefined metrics.
func MakeProtocolViewModel(state *internal.StateModel) ProtocolViewModel {
	return ProtocolViewModel{
		Data:           state.Status,
		TerminalWidth:  0,
		TerminalHeight: 0,
		IsVisible:      true,
	}
}
