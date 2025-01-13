package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProtocolViewModel includes the internal.StatusModel and internal.Model
type ProtocolViewModel struct {
	Data           algod.Status
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
	case algod.Status:
		m.Data = msg
		return m, nil
	// Update Viewport Size
	case tea.WindowSizeMsg:
		m.TerminalWidth = msg.Width
		m.TerminalHeight = msg.Height
		return m, nil
	}
	// Return the updated model to the Bubble Tea runtime for processing.
	return m, nil
}

func formatProtocolVote(status algod.Status) string {
	voting := status.UpgradeYesVotes > 0
	if !voting {
		return "No"
	}
	totalVotesCast := status.UpgradeYesVotes + status.UpgradeNoVotes

	percentageYes := 100 * status.UpgradeYesVotes / totalVotesCast
	percentageProgress := 100 * totalVotesCast / status.UpgradeVotesRequired

	statusString := fmt.Sprintf("Voting %d%% complete, %d%% Yes", percentageProgress, percentageYes)

	passing := status.UpgradeYesVotes > status.UpgradeVotesRequired
	if passing {
		statusString = statusString + ", will pass"
	}
	return statusString
}

// View renders the view for the ProtocolViewModel according to the current state and dimensions.
func (m ProtocolViewModel) View() string {
	if !m.IsVisible {
		return ""
	}
	if m.TerminalWidth <= 0 {
		return "Loading...\n\n\n\n\n\n"
	}
	beginning := style.Blue.Render(" Node: ") + m.Data.Version

	isCompact := m.TerminalWidth < 90

	if isCompact && m.TerminalHeight < 26 {
		return ""
	}

	end := ""
	if m.Data.NeedsUpdate && !isCompact {
		end += style.Green.Render("[UPDATE AVAILABLE] ")
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
	rows = append(rows, style.Blue.Render(" Network: ")+m.Data.Network)
	if !isCompact {
		rows = append(rows, "")
	}
	rows = append(rows, style.Blue.Render(" Protocol Upgrade: ")+formatProtocolVote(m.Data))

	if isCompact && m.Data.NeedsUpdate {
		rows = append(rows, style.Blue.Render(" Upgrade Available: ")+style.Green.Render(strconv.FormatBool(m.Data.NeedsUpdate)))
	}
	return style.WithTitle("Protocol", style.ApplyBorder(max(0, size-2), 5, "5").Render(lipgloss.JoinVertical(lipgloss.Left,
		rows...,
	)))
}

// MakeProtocolViewModel constructs a ProtocolViewModel using a given StatusModel and predefined metrics.
func MakeProtocolViewModel(state *algod.StateModel) ProtocolViewModel {
	return ProtocolViewModel{
		Data:           state.Status,
		TerminalWidth:  0,
		TerminalHeight: 0,
		IsVisible:      true,
	}
}
