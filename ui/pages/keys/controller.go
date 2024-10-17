package keys

import (
	"github.com/algorandfoundation/hack-tui/api"
	"github.com/algorandfoundation/hack-tui/internal"
	"github.com/algorandfoundation/hack-tui/ui/pages"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// EmitKeySelected waits for and retrieves a new set of table rows from a given channel.
func EmitKeySelected(key *api.ParticipationKey) tea.Cmd {
	return func() tea.Msg {
		return key
	}
}

type DeleteKey *api.ParticipationKey

func EmitDeleteKey(key *api.ParticipationKey) tea.Cmd {
	return func() tea.Msg {
		return DeleteKey(key)
	}
}
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
		m.table.SetWidth(msg.Width - lipgloss.Width(pages.Padding1("")) - 4)
		m.table.SetHeight(msg.Height - lipgloss.Height(pages.Padding1("")) - lipgloss.Height(m.controls.View()))
		m.table.SetColumns(m.makeColumns(msg.Width - lipgloss.Width(pages.Padding1("")) - 14))
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)
	m.controls, cmd = m.controls.HandleMessage(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}
