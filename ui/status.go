package ui

import (
	"context"
	"fmt"
	"github.com/algorandfoundation/hack-tui/api"
	"github.com/algorandfoundation/hack-tui/internal"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strconv"
	"strings"
)

// StatusViewModel is extended from the internal.StatusModel
type StatusViewModel struct {
	Status    *internal.StatusModel
	Metrics   *internal.MetricsModel
	ViewWidth int
	IsVisible bool
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
	case uint64:
		m.Status.LastRound = msg
	case tea.WindowSizeMsg:
		m.ViewWidth = msg.Width
	// Is it a key press?
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "h":
			m.IsVisible = !m.IsVisible
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	return m, waitForUint64(m.Status.HeartBeat)
}

// View handles the render cycle
func (m StatusViewModel) View() string {
	if !m.IsVisible {
		return ""
	}

	if m.ViewWidth <= 0 {
		return "Loading...\n\n\n\n\n\n"
	}
	beginning := blue.Render(" Latest Round: ") + strconv.Itoa(int(m.Status.LastRound))
	end := yellow.Render(strings.ToUpper(m.Status.State)) + " "
	middle := strings.Repeat(" ", max(0, m.ViewWidth/2-(lipgloss.Width(beginning)+lipgloss.Width(end)+2)))

	// Last Round
	row1 := lipgloss.JoinHorizontal(lipgloss.Left, beginning, middle, end)

	beginning = blue.Render(" Round time: ") + fmt.Sprintf("%.2fs", m.Metrics.RoundTime)
	end = fmt.Sprintf("%d KB/s ", m.Metrics.TX/1024) + green.Render("TX ")
	middle = strings.Repeat(" ", max(0, m.ViewWidth/2-(lipgloss.Width(beginning)+lipgloss.Width(end)+2)))

	row2 := lipgloss.JoinHorizontal(lipgloss.Left, beginning, middle, end)

	beginning = blue.Render(" TPS:") + fmt.Sprintf("%d", m.Metrics.TPS)
	end = fmt.Sprintf("%d KB/s ", m.Metrics.RX/1024) + green.Render("RX ")
	middle = strings.Repeat(" ", max(0, m.ViewWidth/2-(lipgloss.Width(beginning)+lipgloss.Width(end)+2)))

	row3 := lipgloss.JoinHorizontal(lipgloss.Left, beginning, middle, end)

	return topSections(max(0, m.ViewWidth/2)).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			row1,
			"",
			cyan.Render(" -- 100 round average --"),
			row2,
			row3,
		))
}

// MakeStatusViewModel constructs the model to be used in a tea.Program
func MakeStatusViewModel(ctx context.Context, client *api.ClientWithResponses) (StatusViewModel, error) {
	status := internal.StatusModel{
		HeartBeat:   make(chan uint64),
		LastRound:   0,
		NeedsUpdate: true,
		State:       "SYNCING",
	}
	// Create the Model
	m := StatusViewModel{
		Status: &status,
		Metrics: &internal.MetricsModel{
			RoundTime: 0,
			TX:        0,
			RX:        0,
			TPS:       0,
		},
		ViewWidth: 80,
		IsVisible: true,
	}

	err := m.Status.Fetch(ctx, client)
	if err != nil {
		return m, err
	}

	// Watch for block changes
	go func() {
		err := m.Status.Watch(ctx, client)
		// TODO: Update render and better error handling
		if err != nil {
			panic(err)
		}
	}()
	return m, nil
}
