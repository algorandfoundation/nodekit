package modal

import (
	"fmt"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/algorandfoundation/algorun-tui/internal/algod/participation"
	"github.com/algorandfoundation/algorun-tui/ui/app"
	"github.com/algorandfoundation/algorun-tui/ui/modals/generate"
	"github.com/algorandfoundation/algorun-tui/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"time"
)

// Init initializes the current ViewModel by batching initialization commands for all associated modal ViewModels.
func (m ViewModel) Init() tea.Cmd {
	return tea.Batch(
		m.infoModal.Init(),
		m.exceptionModal.Init(),
		m.transactionModal.Init(),
		m.confirmModal.Init(),
		m.generateModal.Init(),
	)
}

// HandleMessage processes the given message, updates the ViewModel state, and returns any commands to execute.
func (m ViewModel) HandleMessage(msg tea.Msg) (*ViewModel, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	switch msg := msg.(type) {
	case error:
		m.Open = true
		m.exceptionModal.Message = msg.Error()
		m.SetType(app.ExceptionModal)
	case participation.ShortLinkResponse:
		m.Open = true
		m.SetShortLink(msg)
		m.SetType(app.TransactionModal)
	case *algod.StateModel:
		// Clear the catchup modal
		if msg.Status.State != algod.FastCatchupState && m.Type == app.ExceptionModal && m.title == "Fast Catchup" {
			m.Open = false
			m.SetType(app.InfoModal)
		}

		m.State = msg
		m.transactionModal.State = msg
		m.infoModal.State = msg

		if m.State.Status.State == algod.FastCatchupState {
			m.Open = true
			m.SetType(app.ExceptionModal)
			m.exceptionModal.Message = style.LightBlue(lipgloss.JoinVertical(lipgloss.Top,
				"Please wait while your node syncs with the network.",
				"This process can take up to an hour.",
				"",
				fmt.Sprintf("Accounts Processed:   %d / %d", m.State.Status.CatchpointAccountsProcessed, m.State.Status.CatchpointAccountsTotal),
				fmt.Sprintf("Accounts Verified:    %d / %d", m.State.Status.CatchpointAccountsVerified, m.State.Status.CatchpointAccountsTotal),
				fmt.Sprintf("Key Values Processed: %d / %d", m.State.Status.CatchpointKeyValueProcessed, m.State.Status.CatchpointKeyValueTotal),
				fmt.Sprintf("Key Values Verified:  %d / %d", m.State.Status.CatchpointKeyValueVerified, m.State.Status.CatchpointKeyValueTotal),
				fmt.Sprintf("Downloaded blocks:    %d / %d", m.State.Status.CatchpointBlocksAcquired, m.State.Status.CatchpointBlocksTotal),
				"",
				fmt.Sprintf("Sync Time: %ds", m.State.Status.SyncTime/int(time.Second)),
			))
			m.borderColor = "7"
			m.controls = ""
			m.title = "Fast Catchup"

		} else if m.Type == app.TransactionModal && m.transactionModal.Participation != nil {
			acct, ok := msg.Accounts[m.Address]
			// If the previous state is not active
			if ok {
				if !m.transactionModal.Active {
					if acct.Participation != nil &&
						acct.Participation.VoteFirstValid == m.transactionModal.Participation.Key.VoteFirstValid {
						m.SetActive(true)
						m.infoModal.Active = true
						m.SetType(app.InfoModal)
					}
				} else {
					if acct.Participation == nil {
						m.SetActive(false)
						m.infoModal.Active = false
						m.transactionModal.Active = false
						m.SetType(app.InfoModal)
					}
				}
			}

		}

	case app.ModalEvent:
		if msg.Type == app.ExceptionModal {
			m.Open = true
			m.exceptionModal.Message = msg.Err.Error()
			m.generateModal.SetStep(generate.AddressStep)
			m.SetType(app.ExceptionModal)
		}

		if msg.Type == app.InfoModal {
			m.generateModal.SetStep(generate.AddressStep)
		}
		// On closing events
		if msg.Type == app.CloseModal {
			m.Open = false
			m.generateModal.Input.Focus()
		} else {
			m.Open = true
		}
		// When something has triggered a cancel
		if msg.Type == app.CancelModal {
			switch m.Type {
			case app.InfoModal:
				m.Open = false
			case app.GenerateModal:
				m.Open = false
				m.SetType(app.InfoModal)
				m.generateModal.SetStep(generate.AddressStep)
				m.generateModal.Input.Focus()
			case app.TransactionModal:
				m.SetType(app.InfoModal)
			case app.ExceptionModal:
				m.Open = false
			case app.ConfirmModal:
				m.SetType(app.InfoModal)
			}
		}

		if msg.Type != app.CloseModal && msg.Type != app.CancelModal {
			m.SetKey(msg.Key)
			m.SetAddress(msg.Address)
			m.SetActive(msg.Active)
			m.SetType(msg.Type)
		}

	// Handle Modal Type
	case app.ModalType:
		m.SetType(msg)

	// Handle Confirmation Dialog Delete Finished
	case app.DeleteFinished:
		m.Open = false
		m.Type = app.InfoModal
		if msg.Err != nil {
			m.Open = true
			m.Type = app.ExceptionModal
			m.exceptionModal.Message = "Delete failed"
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
		m.exceptionModal, cmd = m.exceptionModal.HandleMessage(modalMsg)
		cmds = append(cmds, cmd)
		return &m, tea.Batch(cmds...)
	}

	// Only trigger modal commands when they are active
	switch m.Type {
	case app.ExceptionModal:
		m.exceptionModal, cmd = m.exceptionModal.HandleMessage(msg)
	case app.InfoModal:
		m.infoModal, cmd = m.infoModal.HandleMessage(msg)
	case app.TransactionModal:
		m.transactionModal, cmd = m.transactionModal.HandleMessage(msg)

	case app.ConfirmModal:
		m.confirmModal, cmd = m.confirmModal.HandleMessage(msg)
	case app.GenerateModal:
		m.generateModal, cmd = m.generateModal.HandleMessage(msg)
	}
	cmds = append(cmds, cmd)

	return &m, tea.Batch(cmds...)
}

// Update processes the given message, updates the ViewModel state, and returns the updated model and accompanying commands.
func (m ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.HandleMessage(msg)
}
