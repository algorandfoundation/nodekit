package accounts

import (
	"github.com/algorandfoundation/algorun-tui/internal"
	"github.com/algorandfoundation/algorun-tui/ui/app"
	"github.com/algorandfoundation/algorun-tui/ui/style"
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
	case *internal.StateModel:
		m.Data = msg
		m.table.SetRows(*m.makeRows())
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			selAcc := m.SelectedAccount()
			if selAcc != nil {
				var cmds []tea.Cmd
				cmds = append(cmds, app.EmitAccountSelected(*selAcc))
				cmds = append(cmds, app.EmitShowPage(app.KeysPage))
				return m, tea.Batch(cmds...)
			}
			return m, nil
		}
	case tea.WindowSizeMsg:
		borderRender := style.Border.Render("")
		borderWidth := lipgloss.Width(borderRender)
		borderHeight := lipgloss.Height(borderRender)

		m.Width = max(0, msg.Width-borderWidth)
		m.Height = max(0, msg.Height-borderHeight)

		m.table.SetWidth(m.Width)
		m.table.SetHeight(max(0, m.Height))
		m.table.SetColumns(m.makeColumns(m.Width))
	}

	// Handle Table Update
	m.table, _ = m.table.Update(msg)

	return m, nil
}
