package ui

import (
	"fmt"
	"github.com/algorandfoundation/hack-tui/internal"
	"github.com/algorandfoundation/hack-tui/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	case internal.StateModel:
		m.Data = &msg
	// Is it a resize event?
	case tea.WindowSizeMsg:
		m.TerminalWidth = msg.Width
		m.TerminalHeight = msg.Height
	// Is it a key press?
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		// always hide on H press
		case "h":
			m.IsVisible = !m.IsVisible
		}
	}
	// Return the updated model to the Bubble Tea runtime for processing.
	return m, nil
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
	end := style.Yellow.Render(strings.ToUpper(m.Data.Status.State)) + " "
	middle := strings.Repeat(" ", max(0, size-(lipgloss.Width(beginning)+lipgloss.Width(end)+2)))

	// Last Round
	row1 := lipgloss.JoinHorizontal(lipgloss.Left, beginning, middle, end)

	beginning = style.Blue.Render(" Round time: ") + fmt.Sprintf("%.2fs", float64(m.Data.Metrics.RoundTime)/float64(time.Second))
	end = fmt.Sprintf("%d KB/s ", m.Data.Metrics.TX/1024) + style.Green.Render("TX ")
	middle = strings.Repeat(" ", max(0, size-(lipgloss.Width(beginning)+lipgloss.Width(end)+2)))

	row2 := lipgloss.JoinHorizontal(lipgloss.Left, beginning, middle, end)

	beginning = style.Blue.Render(" TPS: ") + fmt.Sprintf("%.2f", m.Data.Metrics.TPS)
	end = fmt.Sprintf("%d KB/s ", m.Data.Metrics.RX/1024) + style.Green.Render("RX ")
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
