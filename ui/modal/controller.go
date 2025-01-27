package modal

import (
	"fmt"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/internal/algod/participation"
	"github.com/algorandfoundation/nodekit/ui/app"
	"github.com/algorandfoundation/nodekit/ui/style"
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
func (m ViewModel) HandleMessage(msg tea.Msg) (ViewModel, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	switch msg := msg.(type) {
	// When the state updates
	// TODO: refactor to split this up a bit
	case *algod.StateModel:
		// Clear the catchup modal
		if msg.Status.State != algod.FastCatchupState && m.Type == app.ExceptionModal && m.title == "Fast Catchup" {
			m.Open = false
			m.SetType(app.InfoModal)
		}

		m.State = msg
		m.transactionModal.State = msg
		m.infoModal.State = msg

		// On Fast-Catchup, handle the state as an exception modal
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
			// Return early, skip any checks
			m.exceptionModal, cmd = m.exceptionModal.HandleMessage(msg)
			return m, cmd
		}

		// Get the existing account from the state
		acct, ok := msg.Accounts[m.Address]

		// Handle suspensions
		if ok {
			m.SetSuspended(acct.Participation != nil && acct.Status == "Offline")
		}

		// We found the account, and we are on one of the modals
		if ok && m.Type == app.TransactionModal || m.Type == app.InfoModal {
			// Make sure the transaction modal is set to the current address
			if m.transactionModal.Participation != nil && m.transactionModal.Participation.Address == acct.Address {
				// Actual State
				isOnline := acct.Status == "Online"
				isActive := acct.Participation != nil

				// Derived State
				isValid := isOnline && isActive
				diff, isDifferent, count := participation.HasChanged(*m.transactionModal.Participation, acct.Participation)

				// The account is valid and we registered
				if isValid && !isDifferent && m.Type == app.TransactionModal && !m.transactionModal.OfflineControls {
					m.SetActive(true)
					m.infoModal.Prefix = "Successfully registered online!\n"
					m.HasPrefix = true
					m.SetType(app.InfoModal)
					// For the love of all that is good, please lets refactor this. Preferably with a daemon
				} else if isValid && isDifferent && count != 6 && (m.Type == app.InfoModal || (m.Type == app.TransactionModal && !m.transactionModal.OfflineControls)) {
					// It is online, has a participation key but not the one we are looking at AND all the keys are not different
					// (AND it's the info modal (this case we are checking on enter) OR we are waiting to register a key, and we made a mistake

					// Ahh yes, classic "Set Active to the inverse then only navigate when there is no prefix and it's not the dilution changing"
					// Dilution is likely to match some active keys so we just ignore it. First and last rounds must be unique pairs
					// This is the closest thing we have to state, between this and the transaction modal state it works
					// Set active ensures the offline modal is changed when a corruption happens
					m.SetActive(false)
					if (m.infoModal.Prefix == "" && diff.VoteKeyDilution) || (m.infoModal.Prefix == "" && count <= 4) {
						m.infoModal.Prefix = "***WARNING***\nRegistered online but keys do not fully match\nCheck your registered keys carefully against the node keys\n\n"
						if diff.VoteFirstValid {
							m.infoModal.Prefix = m.infoModal.Prefix + "Mismatched: Vote First Valid\n"
						}
						if diff.VoteLastValid {
							m.infoModal.Prefix = m.infoModal.Prefix + "Mismatched: Vote Last Valid\n"
						}
						if diff.VoteKeyDilution {
							m.infoModal.Prefix = m.infoModal.Prefix + "Mismatched: Vote Key Dilution\n"
						}
						if diff.VoteParticipationKey {
							m.infoModal.Prefix = m.infoModal.Prefix + "Mismatched: Vote Key\n"
						}
						if diff.SelectionParticipationKey {
							m.infoModal.Prefix = m.infoModal.Prefix + "Mismatched: Selection Key\n"
						}
						if diff.StateProofKey {
							m.infoModal.Prefix = m.infoModal.Prefix + "Mismatched: State Proof Key\n"
						}
						m.HasPrefix = true

						m.SetType(app.InfoModal)
					}
				} else if !isOnline && m.Type == app.TransactionModal && m.transactionModal.OfflineControls && m.transactionModal.ATxn.VotePK == nil {
					m.SetActive(false)
					m.infoModal.Prefix = "Successfully registered offline!\n"
					m.HasPrefix = true
					m.SetType(app.InfoModal)

				}
			}
		}

	case app.KeySelectedEvent:
		m.Open = true
		m.SetKey(msg.Key)
		m.SetActive(msg.Active)
		m.SetType(app.InfoModal)
		// Only update prefix when something else has requested it
		if msg.Prefix != "" && m.Type == app.InfoModal {
			m.infoModal.Prefix = msg.Prefix
		}
	case app.OverlayEventType:
		switch msg {
		case app.OverlayEventClose:
			m.Open = false
			m.SetType(app.InfoModal)
			m.generateModal.Reset("")
		case app.OverlayEventCancel:
			switch m.Type {
			case app.InfoModal:
				return m, app.EmitCloseOverlay()
			case app.GenerateModal:
				return m, app.EmitCloseOverlay()
			case app.TransactionModal:
				m.SetType(app.InfoModal)
			case app.ConfirmModal:
				m.SetType(app.InfoModal)
			}
		}
	// Handle Modal Type
	case app.ModalType:
		m.Open = true
		m.SetType(msg)

	case tea.KeyMsg:
		if msg.String() == "q" && m.Type != app.GenerateModal && m.Open {
			return m, tea.Quit
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
		// Exit early and don't apply twice
		cmds = append(cmds, cmd)
		return m, tea.Batch(cmds...)
	}

	// Handle all other messages
	m.confirmModal, cmd = m.confirmModal.HandleMessage(msg)
	cmds = append(cmds, cmd)
	m.infoModal, cmd = m.infoModal.HandleMessage(msg)
	cmds = append(cmds, cmd)
	m.transactionModal, cmd = m.transactionModal.HandleMessage(msg)
	cmds = append(cmds, cmd)
	m.generateModal, cmd = m.generateModal.HandleMessage(msg)
	cmds = append(cmds, cmd)
	m.exceptionModal, cmd = m.exceptionModal.HandleMessage(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// Update processes the given message, updates the ViewModel state, and returns the updated model and accompanying commands.
func (m ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.HandleMessage(msg)
}
