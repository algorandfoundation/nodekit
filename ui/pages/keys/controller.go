package keys

import (
	"github.com/algorandfoundation/hack-tui/internal"
	"github.com/algorandfoundation/hack-tui/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m ViewModel) Init() tea.Cmd {
	return nil
}

func (m ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.HandleMessage(msg)
}

func (m ViewModel) HandleMessage(msg tea.Msg) (ViewModel, tea.Cmd) {
	switch msg := msg.(type) {
	case internal.StateModel:
		m.Data = msg.ParticipationKeys
		m.table.SetRows(m.makeRows(m.Data))
	case internal.Account:
		m.Address = msg.Address
		m.table.SetRows(m.makeRows(m.Data))
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m, EmitKeySelected(m.SelectedKey())
		case "g":
			// TODO: navigation

		case "d":
			return m, EmitDeleteKey(m.SelectedKey())
		case "ctrl+c":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height + 1

		borderRender := style.Border.Render("")
		tableWidth := max(0, msg.Width-lipgloss.Width(borderRender)-20)
		m.table.SetWidth(tableWidth)
		m.table.SetHeight(m.Height - lipgloss.Height(borderRender) - lipgloss.Height(m.controls.View()))
		m.table.SetColumns(m.makeColumns(tableWidth))
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)
	m.controls, cmd = m.controls.HandleMessage(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}
