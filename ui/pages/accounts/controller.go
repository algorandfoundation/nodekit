package accounts

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
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case internal.StateModel:
		m.Data = msg.Accounts
		m.table.SetRows(*m.makeRows())
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			return m, EmitAccountSelected(m.SelectedAccount())
		case "ctrl+c":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		borderRender := style.Border.Render("")
		m.table.SetWidth(max(0, msg.Width-lipgloss.Width(borderRender)))
		m.table.SetHeight(m.Height - lipgloss.Height(borderRender) - lipgloss.Height(m.controls.View()))
		m.table.SetColumns(m.makeColumns(msg.Width - lipgloss.Width(borderRender)))
	}

	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)
	m.controls, cmd = m.controls.HandleMessage(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}
