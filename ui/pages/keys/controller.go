package keys

import (
	"github.com/algorandfoundation/hack-tui/internal"
	"github.com/algorandfoundation/hack-tui/ui/app"
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
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	switch msg := msg.(type) {
	// When the State changes
	case internal.StateModel:
		m.Data = msg.ParticipationKeys
		m.table.SetRows(m.makeRows(m.Data))
	// When the Account is Selected
	case app.AccountSelected:
		m.Address = msg.Address
		m.table.SetRows(m.makeRows(m.Data))
	// When a confirmation Modal is finished deleting
	case app.DeleteFinished:
		if msg.Err != nil {
			panic(msg.Err)
		}
		internal.RemovePartKeyByID(m.Data, msg.Id)
		m.table.SetRows(m.makeRows(m.Data))
	// When the user interacts with the render
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			return m, app.EmitShowPage(app.AccountsPage)
		// Show the Info Modal
		case "enter":
			selKey := m.SelectedKey()
			if selKey != nil {
				// Show the Info Modal with the selected Key
				return m, app.EmitModalEvent(app.ModalEvent{
					Key:     selKey,
					Address: selKey.Address,
					Type:    app.InfoModal,
				})
			}
			return m, nil
		}

	// Handle Resize Events
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

	// Handle Table Update
	m.table, cmd = m.table.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	cmds = append(cmds, cmd)

	// Batch all commands
	return m, tea.Batch(cmds...)
}
