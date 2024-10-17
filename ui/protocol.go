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
	tea.Model
	Status    *internal.StatusModel
	Metrics   *internal.MetricsModel
	ViewWidth int
	IsVisible bool
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
	// Update Viewport Size
	case tea.WindowSizeMsg:
		m.ViewWidth = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		// The H key should hide the render
		case "h":
			m.IsVisible = !m.IsVisible
		case "q", "ctrl+c":
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
	if m.ViewWidth <= 0 {
		return "Loading...\n\n\n\n\n\n"
	}
	beginning := blue.Render(" Node: ") + m.Status.Version
	end := ""
	if m.Status.NeedsUpdate {
		end = green.Render("[UPDATE AVAILABLE] ")
	}

	middle := strings.Repeat(" ", max(0, m.ViewWidth/2-(lipgloss.Width(beginning)+lipgloss.Width(end)+2)))

	// Last Round
	row1 := lipgloss.JoinHorizontal(lipgloss.Left, beginning, middle, end)

	row2 := blue.Render(" Network: ") + m.Status.Network

	row3 := blue.Render(" Protocol Voting: ") + strconv.FormatBool(m.Status.Voting)

	return topSections(max(0, m.ViewWidth/2)).Render(lipgloss.JoinVertical(lipgloss.Left,
		row1,
		"",
		row2,
		"",
		row3,
	))
}

// MakeProtocolViewModel constructs a ProtocolViewModel using a given StatusModel and predefined metrics.
func MakeProtocolViewModel(status *internal.StatusModel) ProtocolViewModel {
	metrics := internal.MetricsModel{
		RoundTime: 2.87,
		TPS:       55,
		RX:        82 * 1024,
		TX:        205 * 1024,
	}

	return ProtocolViewModel{
		Status:    status,
		Metrics:   &metrics,
		ViewWidth: 0,
		IsVisible: true,
	}
}
