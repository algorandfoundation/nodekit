package ui

import (
	"context"
	"fmt"
	"github.com/algorandfoundation/hack-tui/api"
	"github.com/algorandfoundation/hack-tui/internal"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strconv"
	"strings"
)

const useHighPerformanceRenderer = false

var (
	rounderBorder = func() lipgloss.Style {
		return lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder())
	}()
	topSections = func(width int) lipgloss.Style {
		return rounderBorder.
			Width(width - 2).
			Padding(0).
			Margin(0).
			Height(5).
			//BorderBackground(lipgloss.Color("4")).
			BorderForeground(lipgloss.Color("5"))
	}
	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		b.Right = "├"
		return rounderBorder.BorderStyle(b)
	}()
	blue = func() lipgloss.Style {
		return lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	}()
	cyan   = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	yellow = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	green  = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
)

type ViewportViewModel struct {
	ready       bool
	HelpVisible bool
	viewport    viewport.Model
	table       table.Model
	status      *internal.StatusModel
}

func (m ViewportViewModel) Init() tea.Cmd {
	return nil
}

func (m ViewportViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case uint64:
		m.status.LastRound = msg
	case tea.KeyMsg:
		switch msg.String() {
		case "h":
			m.HelpVisible = !m.HelpVisible
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			if m.table.SelectedRow() != nil {
				return m, tea.Batch(
					tea.Printf("Let's go to %s!", m.table.SelectedRow()[1]),
				)
			}

		}

	case tea.WindowSizeMsg:
		headerHeight := lipgloss.Height(m.headerView())
		footerHeight := lipgloss.Height(m.footerView())
		verticalMarginHeight := headerHeight + footerHeight

		// On first run, configure the models
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-verticalMarginHeight)
			m.viewport.YPosition = headerHeight
			m.viewport.HighPerformanceRendering = useHighPerformanceRenderer
			m.viewport.SetContent(m.table.View())
			m.ready = true

			// TODO: Better reactivity and hidden attributes
			fillSize := max(0, (msg.Width-49)/2)
			columns := []table.Column{
				{Title: "Account", Width: fillSize},
				{Title: "Status", Width: hidden(20, fillSize)},
				{Title: "Keys", Width: 4},
				{Title: "Expires", Width: 15},
				{Title: "Last Used", Width: 10},
				{Title: "Balance", Width: fillSize},
			}

			rows := []table.Row{
				{"QNZ7GONNHTNXFW56Y24CNJQEMYKZKKI566ASNSWPD24VSGKJWHGO6QOP7U", "Active", "4", "42 days", "NA", "42,000 ALGO"},
				{"WZ7BQUYLGP5GCWVHH6PJJCGCIHRV4K7ZDFWHED74HGLUCB3GTDVPNFRVUM", "Cooldown (31 rounds)", "1", "169 days", "NA", "13,000 ALGO"},
			}

			m.table = table.New(
				table.WithColumns(columns),
				table.WithRows(rows),
				table.WithFocused(true),
				table.WithHeight(m.viewport.Height-verticalMarginHeight),
			)

			s := table.DefaultStyles()
			s.Header = s.Header.
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				BorderBottom(true).
				Bold(false)
			s.Selected = s.Selected.
				Foreground(lipgloss.Color("229")).
				Background(lipgloss.Color("57")).
				Bold(false)
			m.table.SetStyles(s)
			m.viewport.YPosition = headerHeight + 1
		} else { // Run the update cycle
			m.table.SetWidth(msg.Width)
			m.table.SetHeight(msg.Height - verticalMarginHeight)

			fillSize := (msg.Width - 62) / 2
			columns := []table.Column{
				{Title: "Account", Width: fillSize},
				{Title: "Status", Width: hidden(20, fillSize)},
				{Title: "Keys", Width: 4},
				{Title: "Expires", Width: 15},
				{Title: "Last Used", Width: 10},
				{Title: "Balance", Width: fillSize},
			}
			m.table.SetColumns(columns)

			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - verticalMarginHeight
		}

		if useHighPerformanceRenderer {
			cmds = append(cmds, viewport.Sync(m.viewport))
		}
	}

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)

	cmds = append(cmds, waitForUint64(m.status.HeartBeat))
	return m, tea.Batch(cmds...)
}

// hidden returns 0 when the width is greater than the fill
func hidden(width int, fillSize int) int {
	if fillSize < width {
		return 0
	}
	return width
}

func (m ViewportViewModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}
	m.viewport.SetContent(m.table.View())
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.footerView())
}

// TODO: Move to internal Metrics Model
type MetricsModel struct {
	LastRound int
	RoundTime float64
	TPS       int
	State     string
	RX        int
	TX        int
}

func (m ViewportViewModel) metricsView(status MetricsModel) string {
	if m.viewport.Width <= 0 {
		return ""
	}
	beginning := blue.Render(" Latest Round: ") + strconv.Itoa(status.LastRound)
	end := yellow.Render(strings.ToUpper(status.State)) + " "
	middle := strings.Repeat(" ", max(0, m.viewport.Width/2-(lipgloss.Width(beginning)+lipgloss.Width(end)+2)))

	// Last Round
	row1 := lipgloss.JoinHorizontal(lipgloss.Left, beginning, middle, end)

	beginning = blue.Render(" Round time: ") + fmt.Sprintf("%.2fs", status.RoundTime)
	end = fmt.Sprintf("%d KB/s ", status.TX/1024) + green.Render("TX ")
	middle = strings.Repeat(" ", max(0, m.viewport.Width/2-(lipgloss.Width(beginning)+lipgloss.Width(end)+2)))

	row2 := lipgloss.JoinHorizontal(lipgloss.Left, beginning, middle, end)

	beginning = blue.Render(" TPS:") + fmt.Sprintf("%d", status.TPS)
	end = fmt.Sprintf("%d KB/s ", status.RX/1024) + green.Render("RX ")
	middle = strings.Repeat(" ", max(0, m.viewport.Width/2-(lipgloss.Width(beginning)+lipgloss.Width(end)+2)))

	row3 := lipgloss.JoinHorizontal(lipgloss.Left, beginning, middle, end)
	return lipgloss.JoinVertical(lipgloss.Left,
		row1,
		"",
		cyan.Render(" -- 100 round average --"),
		row2,
		row3,
	)
}

func (m ViewportViewModel) consensusView(status internal.StatusModel) string {
	if m.viewport.Width <= 0 {
		return ""
	}
	beginning := blue.Render(" Node: ") + status.Version
	end := ""
	if status.NeedsUpdate {
		end = green.Render("[UPDATE AVAILABLE] ")
	}

	middle := strings.Repeat(" ", max(0, m.viewport.Width/2-(lipgloss.Width(beginning)+lipgloss.Width(end)+2)))

	// Last Round
	row1 := lipgloss.JoinHorizontal(lipgloss.Left, beginning, middle, end)

	row2 := blue.Render(" Network: ") + status.Network

	row3 := blue.Render(" Protocol Voting: ") + strconv.FormatBool(status.Voting)

	return lipgloss.JoinVertical(lipgloss.Left,
		row1,
		"",
		row2,
		"",
		row3,
	)
}
func (m ViewportViewModel) headerView() string {
	if !m.HelpVisible {
		return ""
	}
	metrics := MetricsModel{
		LastRound: int(m.status.LastRound),
		RoundTime: 2.87,
		TPS:       55,
		State:     "syncing",
		RX:        82 * 1024,
		TX:        205 * 1024,
	}
	left := topSections(max(0, m.viewport.Width/2)).Render(m.metricsView(metrics))

	right := topSections(max(0, m.viewport.Width/2)).Render(m.consensusView(*m.status))

	return lipgloss.JoinHorizontal(lipgloss.Center, left, right)
}

func (m ViewportViewModel) footerView() string {
	if !m.HelpVisible {
		return ""
	}
	info := infoStyle.Render(" (q)uit | (d)elete | (g)enerate | (h)ide ")
	difference := m.viewport.Width - lipgloss.Width(info)

	line := strings.Repeat("─", max(0, difference/2))
	return lipgloss.JoinHorizontal(lipgloss.Center, line, info, line)
}

func MakeViewportViewModel(ctx context.Context, client *api.ClientWithResponses) (*ViewportViewModel, error) {

	status := internal.StatusModel{
		HeartBeat:   make(chan uint64),
		Voting:      false,
		NeedsUpdate: true,
	}
	err := status.Fetch(ctx, client)
	if err != nil {
		return nil, err
	}
	m := ViewportViewModel{
		HelpVisible: true,
		status:      &status,
	}

	// Watch for block changes
	go func() {
		err := status.Watch(ctx, client)
		// TODO: Update render and better error handling
		if err != nil {
			panic(err)
		}
	}()
	return &m, nil
}
