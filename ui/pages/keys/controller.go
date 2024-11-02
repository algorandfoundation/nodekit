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
	case DeleteFinished:
		if m.SelectedKeyToDelete == nil {
			panic("SelectedKeyToDelete is unexpectedly nil")
		}
		internal.RemovePartKeyByID(m.Data, m.SelectedKeyToDelete.Id)
		m.SelectedKeyToDelete = nil
		m.table.SetRows(m.makeRows(m.Data))

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			selKey := m.SelectedKey()
			if selKey != nil {
				return m, EmitKeySelected(selKey)
			}
			return m, nil
		case "g":
			// TODO: navigation
		case "d":
			if m.SelectedKeyToDelete == nil {
				m.SelectedKeyToDelete = m.SelectedKey()
			} else {
				m.SelectedKeyToDelete = nil
			}
			return m, nil
		case "y": // "Yes do delete" option in the delete confirmation modal
			if m.SelectedKeyToDelete != nil {
				return m, EmitDeleteKey(m.SelectedKeyToDelete)
			}
			return m, nil
		case "n": // "do NOT delete" option in the delete confirmation modal
			if m.SelectedKeyToDelete != nil {
				m.SelectedKeyToDelete = nil
			}
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		borderRender := style.Border.Render("")
		borderWidth := lipgloss.Width(borderRender)
		borderHeight := lipgloss.Height(borderRender)

		m.Width = max(0, msg.Width-borderWidth)
		m.Height = max(0, msg.Height-borderHeight)
		m.table.SetWidth(m.Width)
		m.table.SetHeight(m.Height)
		m.table.SetColumns(m.makeColumns(m.Width))
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}
