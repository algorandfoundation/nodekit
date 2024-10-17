package ui

import (
	"context"
	"fmt"
	"github.com/algorandfoundation/hack-tui/api"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const useHighPerformanceRenderer = false

type ViewportViewModel struct {
	// Status Component
	status StatusViewModel
	// Protocol Component
	protocol ProtocolViewModel
	// Application Controls
	controls ControlViewModel

	ready    bool
	viewport viewport.Model
	// TODO: move to custom component
	table table.Model
}

// Init is a no-op
func (m ViewportViewModel) Init() tea.Cmd {
	return nil
}

// Update Handle the viewport lifecycle
func (m ViewportViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case uint64:
		m.status.Status.LastRound = msg
	case tea.KeyMsg:
		switch msg.String() {
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
		footerHeight := lipgloss.Height(m.controls.View())
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
	m.controls, cmd = m.controls.HandleMessage(msg)
	cmds = append(cmds, cmd)
	m.protocol, cmd = m.protocol.HandleMessage(msg)
	cmds = append(cmds, cmd)
	m.status, cmd = m.status.HandleMessage(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// View renders the viewport.Model
func (m ViewportViewModel) View() string {
	if !m.ready {
		return "\n  Initializing..."
	}
	m.viewport.SetContent(m.table.View())
	return fmt.Sprintf("%s\n%s\n%s", m.headerView(), m.viewport.View(), m.controls.View())
}

// headerView generates the top elements
func (m ViewportViewModel) headerView() string {
	// TODO: Stack Vertically on small screens
	render := lipgloss.JoinHorizontal(lipgloss.Center, m.status.View(), m.protocol.View())
	return render
}

// MakeViewportViewModel handles the construction of the TUI viewport
func MakeViewportViewModel(ctx context.Context, client *api.ClientWithResponses) (*ViewportViewModel, error) {
	controls := MakeControlViewModel()

	status, err := MakeStatusViewModel(ctx, client)
	if err != nil {
		return nil, err
	}
	protocol := MakeProtocolViewModel(status.Status)

	m := ViewportViewModel{
		status:   status,
		protocol: protocol,
		controls: controls,
	}

	return &m, nil
}
