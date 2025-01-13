package ui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProtocolViewModel includes the internal.StatusModel and internal.Model
type ProtocolViewModel struct {
	Data           algod.Status
	Metrics        algod.Metrics
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
	case *algod.StateModel:
		m.Data = msg.Status
		m.Metrics = msg.Metrics
	case algod.Status:
		m.Data = msg
		return m, nil
	case algod.Metrics:
		m.Metrics = msg
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

func plural(singularForm string, value int) string {
	if value == 1 {
		return singularForm
	} else {
		return singularForm + "s"
	}
}

func formatScheduledUpgrade(status algod.Status, metrics algod.Metrics) string {
	roundDelta := status.NextVersionRound - int(status.LastRound)
	eta := time.Duration(roundDelta) * metrics.RoundTime
	minutes := int(eta.Minutes()) % 60
	hours := int(eta.Hours()) % 24
	days := int(eta.Hours()) / 24
	str := "Scheduled"
	if days > 0 {
		str = str + fmt.Sprintf(" %d %s", days, plural("day", days))
	}
	if hours > 0 {
		str = str + fmt.Sprintf(" %d %s", hours, plural("hour", hours))
	}
	if days == 0 && minutes > 0 {
		str = str + fmt.Sprintf(" %d %s", minutes, plural("min", minutes))
	}
	return str
}

func formatProtocolVote(status algod.Status, metrics algod.Metrics) string {
	if status.NextVersionRound > int(status.LastRound)+1 {
		return formatScheduledUpgrade(status, metrics)
	}

	voting := status.UpgradeYesVotes > 0 || status.UpgradeNoVotes > 0
	if !voting {
		return "No"
	}

	totalVotesCast := status.UpgradeYesVotes + status.UpgradeNoVotes
	percentageProgress := 100 * totalVotesCast / status.UpgradeVoteRounds
	percentageYes := 100 * status.UpgradeYesVotes / totalVotesCast

	label := "Yes"
	percentageVoteDisplay := percentageYes
	if percentageYes < 50 {
		label = "No"
		percentageVoteDisplay = 100 * status.UpgradeNoVotes / totalVotesCast
	}
	statusString := fmt.Sprintf("Voting %d%% complete, %d%% %s", percentageProgress, percentageVoteDisplay, label)

	passing := status.UpgradeYesVotes > status.UpgradeVotesRequired
	if passing {
		statusString = statusString + ", will pass"
	}
	failThreshold := status.UpgradeVoteRounds - status.UpgradeVotesRequired
	if status.UpgradeNoVotes > failThreshold {
		statusString = statusString + ", will fail"
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
	rows = append(rows, style.Blue.Render(" Protocol Upgrade: ")+formatProtocolVote(m.Data, m.Metrics))
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
		Metrics:        state.Metrics,
		TerminalWidth:  0,
		TerminalHeight: 0,
		IsVisible:      true,
	}
}
