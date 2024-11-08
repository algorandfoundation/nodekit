package modal

import (
	"github.com/algorandfoundation/hack-tui/ui/modals/confirm"
	"github.com/algorandfoundation/hack-tui/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m ViewModel) Init() tea.Cmd {
	return nil
}
func (m ViewModel) HandleMessage(msg tea.Msg) (*ViewModel, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	switch msg := msg.(type) {
	case error:
		m.Open = true
		m.Page = ExceptionModal
		m.exceptionModal.Message = msg.Error()
	// Handle Confirmation Dialog Cancel
	case confirm.Msg:
		if msg == nil {
			m.Page = InfoModal
		}
	// Handle Confirmation Dialog Delete Finished
	case confirm.DeleteFinished:
		m.Open = false
		m.Page = InfoModal

	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			switch m.Page {
			case InfoModal:
				m.Open = false
			case GenerateModal:
				m.Open = false
				m.Page = InfoModal
			case TransactionModal:
				m.Page = InfoModal
			case ExceptionModal:
				m.Open = false
			case ConfirmModal:
				m.Page = InfoModal
			}
		case "g":
			if m.Page != GenerateModal {
				m.Page = GenerateModal
			}
		case "d":
			if m.Page == InfoModal {
				m.Page = ConfirmModal
			}
		case "o":
			if m.Page == InfoModal {
				m.Page = TransactionModal
				m.transactionModal.UpdateState()
			}
		case "enter":
			if m.Page == InfoModal {
				m.Page = TransactionModal
			}
			if m.Page == TransactionModal {
				m.Page = InfoModal
			}
		}
	// Handle View Size changes
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height

		b := style.Border.Render("")
		// Custom size message
		modalMsg := tea.WindowSizeMsg{
			Width:  m.Width - lipgloss.Width(b),
			Height: m.Height - lipgloss.Height(b),
		}

		// Handle the page resize event
		m.infoModal, cmd = m.infoModal.HandleMessage(modalMsg)
		cmds = append(cmds, cmd)
		m.transactionModal, cmd = m.transactionModal.HandleMessage(modalMsg)
		cmds = append(cmds, cmd)
		m.confirmModal, cmd = m.confirmModal.HandleMessage(modalMsg)
		cmds = append(cmds, cmd)
		m.generateModal, cmd = m.generateModal.HandleMessage(modalMsg)
		cmds = append(cmds, cmd)
		return &m, tea.Batch(cmds...)
	}

	// Only trigger modal commands when they are active
	switch m.Page {
	case ExceptionModal:
		m.exceptionModal, cmd = m.exceptionModal.HandleMessage(msg)
		m.title = m.exceptionModal.Title
		m.controls = m.exceptionModal.Controls
		m.borderColor = m.exceptionModal.BorderColor
	case InfoModal:
		m.infoModal, cmd = m.infoModal.HandleMessage(msg)
		m.title = m.infoModal.Title
		m.controls = m.infoModal.Controls
		m.borderColor = m.infoModal.BorderColor
	case TransactionModal:
		m.transactionModal, cmd = m.transactionModal.HandleMessage(msg)
		m.title = m.transactionModal.Title
		m.controls = m.transactionModal.Controls
		m.borderColor = m.transactionModal.BorderColor
	case ConfirmModal:
		m.confirmModal, cmd = m.confirmModal.HandleMessage(msg)
		m.title = m.confirmModal.Title
		m.controls = m.confirmModal.Controls
		m.borderColor = m.confirmModal.BorderColor
	case GenerateModal:
		m.generateModal, cmd = m.generateModal.HandleMessage(msg)
		m.title = m.generateModal.Title
		m.controls = m.generateModal.Controls
		m.borderColor = m.generateModal.BorderColor
	}
	cmds = append(cmds, cmd)
	return &m, tea.Batch(cmds...)
}
func (m ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.HandleMessage(msg)
}
