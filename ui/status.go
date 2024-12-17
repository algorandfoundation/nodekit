package ui

import (
	"fmt"
	"github.com/algorandfoundation/algorun-tui/internal"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/algorandfoundation/algorun-tui/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"math"
	"strconv"
	"strings"
	"time"
)

// StatusViewModel is extended from the internal.StatusModel
type StatusViewModel struct {
	Data           *internal.StateModel
	TerminalWidth  int
	TerminalHeight int
	IsVisible      bool
}

// Init has no I/O right now
func (m StatusViewModel) Init() tea.Cmd {
	return nil
}

// Update is called when the user interacts with the render
func (m StatusViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.HandleMessage(msg)
}

// HandleMessage is called when the user interacts with the render
func (m StatusViewModel) HandleMessage(msg tea.Msg) (StatusViewModel, tea.Cmd) {
	switch msg := msg.(type) {
	// Is it a heartbeat of the latest round?
	case *internal.StateModel:
		m.Data = msg
	// Is it a resize event?
	case tea.WindowSizeMsg:
		m.TerminalWidth = msg.Width
		m.TerminalHeight = msg.Height
	}
	// Return the updated model to the Bubble Tea runtime for processing.
	return m, nil
}

// getBitRate converts a given byte rate to a human-readable string format. The output may vary from B/s to GB/s.
func getBitRate(bytes int) string {
	txString := fmt.Sprintf("%d B/s ", bytes)
	if bytes >= 1024 {
		txString = fmt.Sprintf("%d KB/s ", bytes/(1<<10))
	}
	if bytes >= int(math.Pow(1024, 2)) {
		txString = fmt.Sprintf("%d MB/s ", bytes/(1<<20))
	}
	if bytes >= int(math.Pow(1024, 3)) {
		txString = fmt.Sprintf("%d GB/s ", bytes/(1<<30))
	}

	return txString
}

// View handles the render cycle
func (m StatusViewModel) View() string {
	if !m.IsVisible {
		return ""
	}

	if m.TerminalWidth <= 0 {
		return "Loading...\n\n\n\n\n\n"
	}

	isCompact := m.TerminalWidth < 90

	var size int
	if isCompact {
		size = m.TerminalWidth
	} else {
		size = m.TerminalWidth / 2
	}
	beginning := style.Blue.Render(" Latest Round: ") + strconv.Itoa(int(m.Data.Status.LastRound))

	var end string
	switch m.Data.Status.State {
	case algod.StableState:
		end = style.Green.Render(strings.ToUpper(string(m.Data.Status.State))) + " "
	default:
		end = style.Yellow.Render(strings.ToUpper(string(m.Data.Status.State))) + " "
	}
	middle := strings.Repeat(" ", max(0, size-(lipgloss.Width(beginning)+lipgloss.Width(end)+2)))

	// Last Round
	row1 := lipgloss.JoinHorizontal(lipgloss.Left, beginning, middle, end)

	roundTime := fmt.Sprintf("%.2fs", float64(m.Data.Metrics.RoundTime)/float64(time.Second))
	if m.Data.Status.State != algod.StableState {
		roundTime = "--"
	}
	beginning = style.Blue.Render(" Round time: ") + roundTime
	end = getBitRate(m.Data.Metrics.TX) + style.Green.Render("TX ")
	middle = strings.Repeat(" ", max(0, size-(lipgloss.Width(beginning)+lipgloss.Width(end)+2)))

	row2 := lipgloss.JoinHorizontal(lipgloss.Left, beginning, middle, end)

	tps := fmt.Sprintf("%.2f", m.Data.Metrics.TPS)
	if m.Data.Status.State != algod.StableState {
		tps = "--"
	}
	beginning = style.Blue.Render(" TPS: ") + tps
	end = getBitRate(m.Data.Metrics.RX) + style.Green.Render("RX ")
	middle = strings.Repeat(" ", max(0, size-(lipgloss.Width(beginning)+lipgloss.Width(end)+2)))

	row3 := lipgloss.JoinHorizontal(lipgloss.Left, beginning, middle, end)

	return style.WithTitle("Status", style.ApplyBorder(max(0, size-2), 5, "5").Render(
		lipgloss.JoinVertical(lipgloss.Left,
			row1,
			"",
			style.Cyan.Render(" -- "+strconv.Itoa(m.Data.Metrics.Window)+" round average --"),
			row2,
			row3,
		)))
}

// MakeStatusViewModel constructs the model to be used in a tea.Program
func MakeStatusViewModel(state *internal.StateModel) StatusViewModel {
	// Create the Model
	m := StatusViewModel{
		Data:          state,
		TerminalWidth: 80,
		IsVisible:     true,
	}
	return m
}
