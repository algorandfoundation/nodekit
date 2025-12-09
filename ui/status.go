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

// StatusViewModel is extended from the internal.StatusModel
type StatusViewModel struct {
	Data           *algod.StateModel
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
	case *algod.StateModel:
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
func getBitRate(bytes uint64) string {
	txString := fmt.Sprintf("%d B/s", bytes)
	if bytes >= 1024 {
		txString = fmt.Sprintf("%d KB/s", bytes/(1<<10))
	}
	if bytes >= uint64(float64(1024*1024)) {
		txString = fmt.Sprintf("%d MB/s", bytes/(1<<20))
	}
	if bytes >= uint64(float64(1024*1024*1024)) {
		txString = fmt.Sprintf("%d GB/s", bytes/(1<<30))
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
	isP2PHybridEnabled := m.Data.Config != nil && m.Data.Config.EnableP2PHybridMode != nil && *m.Data.Config.EnableP2PHybridMode
	isP2PEnabled := m.Data.Config != nil && m.Data.Config.EnableP2P != nil && *m.Data.Config.EnableP2P && !isP2PHybridEnabled

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

	if isP2PHybridEnabled {
		end = "P2P: " + style.Green.Render("HYBRID") + " "
	} else if isP2PEnabled {
		end = "P2P: " + style.Green.Render("ONLY") + " "
	} else {
		end = "P2P: " + style.Red.Render("NO") + " "
	}
	beginning = ""
	middle = strings.Repeat(" ", max(0, size-(lipgloss.Width(beginning)+lipgloss.Width(end)+2)))
	row2 := lipgloss.JoinHorizontal(lipgloss.Left, beginning, middle, end)

	beginning = style.Cyan.Render(" -- " + strconv.Itoa(m.Data.Metrics.Window) + " round average --")
	// Check metrics to confirm config
	hasWSData := (m.Data.Metrics.TX != 0 || m.Data.Metrics.RX != 0)
	hasP2PData := (m.Data.Metrics.TXP2P != 0 || m.Data.Metrics.RXP2P != 0)
	hasSomeData := hasWSData || hasP2PData
	if isP2PHybridEnabled && hasSomeData && (!hasP2PData || !hasWSData) {
		// Should be P2P and WS
		end = style.Red.Render("Network/Config Mismatch") + " "
	} else if isP2PEnabled && hasSomeData && (!hasP2PData || hasWSData) {
		// Should be ONLY P2P
		end = style.Red.Render("Network/Config Mismatch") + " "
	} else if (!isP2PHybridEnabled && !isP2PEnabled) && hasSomeData && (!hasWSData || hasP2PData) {
		// Should be ONLY WS
		end = style.Red.Render("Network/Config Mismatch") + " "
	} else {
		// Otherwise show peer count
		end = "Peers: "
		if isP2PHybridEnabled {
			end += fmt.Sprintf(" % 4d WS | % 4d P2P ", m.Data.Metrics.PeersWS, m.Data.Metrics.PeersP2P)
		} else if isP2PEnabled {
			end += fmt.Sprintf("%d ", m.Data.Metrics.PeersP2P)
		} else {
			end += fmt.Sprintf("%d ", m.Data.Metrics.PeersWS)
		}
	}
	middle = strings.Repeat(" ", max(0, size-(lipgloss.Width(beginning)+lipgloss.Width(end)+2)))
	row3 := lipgloss.JoinHorizontal(lipgloss.Left, beginning, middle, end)

	roundTime := fmt.Sprintf("%.2fs", float64(m.Data.Metrics.RoundTime)/float64(time.Second))
	if m.Data.Status.State != algod.StableState {
		roundTime = "--"
	}
	beginning = style.Blue.Render(" Round time: ") + roundTime
	end = "Tx: "
	if isP2PHybridEnabled {
		end += fmt.Sprintf("% 8s | % 8s ", getBitRate(m.Data.Metrics.TX), getBitRate(m.Data.Metrics.TXP2P))
	} else if isP2PEnabled {
		end += fmt.Sprintf("%s ", getBitRate(m.Data.Metrics.TXP2P))
	} else {
		end += fmt.Sprintf("%s ", getBitRate(m.Data.Metrics.TX))
	}
	middle = strings.Repeat(" ", max(0, size-(lipgloss.Width(beginning)+lipgloss.Width(end)+2)))
	row4 := lipgloss.JoinHorizontal(lipgloss.Left, beginning, middle, end)

	tps := fmt.Sprintf("%.2f", m.Data.Metrics.TPS)
	if m.Data.Status.State != algod.StableState {
		tps = "--"
	}
	beginning = style.Blue.Render(" TPS: ") + tps
	end = "Rx: "
	if isP2PHybridEnabled {
		end += fmt.Sprintf("% 8s | % 8s ", getBitRate(m.Data.Metrics.RX), getBitRate(m.Data.Metrics.RXP2P))
	} else if isP2PEnabled {
		end += fmt.Sprintf("%s ", getBitRate(m.Data.Metrics.RXP2P))
	} else {
		end += fmt.Sprintf("%s ", getBitRate(m.Data.Metrics.RX))
	}
	middle = strings.Repeat(" ", max(0, size-(lipgloss.Width(beginning)+lipgloss.Width(end)+2)))
	row5 := lipgloss.JoinHorizontal(lipgloss.Left, beginning, middle, end)

	return style.WithTitles(
		"( "+style.Red.Render(fmt.Sprintf("Nodekit-%s", m.Data.Version))+" )",
		lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Render("Status"),
		style.ApplyBorder(max(0, size-2), 5, "5").Render(
			lipgloss.JoinVertical(lipgloss.Left,
				row1,
				row2,
				row3,
				row4,
				row5,
			),
		),
	)
}

// MakeStatusViewModel constructs the model to be used in a tea.Program
func MakeStatusViewModel(state *algod.StateModel) StatusViewModel {
	// Create the Model
	m := StatusViewModel{
		Data:          state,
		TerminalWidth: 80,
		IsVisible:     true,
	}
	return m
}
