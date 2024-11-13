package ui

import (
	"fmt"
	"github.com/algorandfoundation/hack-tui/api"
	"github.com/algorandfoundation/hack-tui/internal"
	"github.com/algorandfoundation/hack-tui/ui/app"
	"github.com/algorandfoundation/hack-tui/ui/modal"
	"github.com/algorandfoundation/hack-tui/ui/modals/exception"
	"github.com/algorandfoundation/hack-tui/ui/pages/accounts"
	"github.com/algorandfoundation/hack-tui/ui/pages/keys"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ViewportViewModel represents the state and view model for a viewport in the application.
type ViewportViewModel struct {
	PageWidth, PageHeight         int
	TerminalWidth, TerminalHeight int

	Data *internal.StateModel

	// Header Components
	status   StatusViewModel
	protocol ProtocolViewModel

	// Pages
	accountsPage accounts.ViewModel
	keysPage     keys.ViewModel

	modal  *modal.ViewModel
	page   app.Page
	client *api.ClientWithResponses

	// Error Handler
	errorMsg  *string
	errorPage *exception.ViewModel
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
	// Handle Header Updates
	m.protocol, cmd = m.protocol.HandleMessage(msg)
	cmds = append(cmds, cmd)
	m.status, cmd = m.status.HandleMessage(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case app.Page:
		if msg == app.KeysPage {
			m.keysPage.Address = m.accountsPage.SelectedAccount().Address
		}
		m.page = msg
	// When the state updates
	case internal.StateModel:
		if m.errorMsg != nil {
			m.errorMsg = nil
			m.page = app.AccountsPage
		}
		m.Data = &msg
	case tea.KeyMsg:
		switch msg.String() {
		case "g":
			// Only open modal when it is closed and not syncing
			if !m.modal.Open && m.Data.Status.State != internal.SyncingState {
				return m, app.EmitModalEvent(app.ModalEvent{
					Key:     nil,
					Address: m.accountsPage.SelectedAccount().Address,
					Type:    app.GenerateModal,
				})
			}

		case "left":
			// Disable when overlay is active or on Accounts
			if m.modal.Open || m.page == app.AccountsPage {
				return m, nil
			}
			// Navigate to the Keys Page
			if m.page == app.KeysPage {
				return m, app.EmitShowPage(app.AccountsPage)
			}
		case "right":
			// Disable when overlay is active
			if m.modal.Open {
				return m, nil
			}
			if m.page == app.AccountsPage {
				selAcc := m.accountsPage.SelectedAccount()
				if selAcc != (internal.Account{}) {
					m.page = app.KeysPage
					return m, app.EmitAccountSelected(selAcc)
				}
				return m, nil
			}
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.TerminalWidth = msg.Width
		m.TerminalHeight = msg.Height
		m.PageWidth = msg.Width
		m.PageHeight = max(0, msg.Height-lipgloss.Height(m.headerView())-1)

		modalMsg := tea.WindowSizeMsg{
			Width:  m.PageWidth - 2,
			Height: m.PageHeight,
		}

		m.modal, cmd = m.modal.HandleMessage(modalMsg)
		cmds = append(cmds, cmd)

		// Custom size message
		pageMsg := tea.WindowSizeMsg{
			Height: m.PageHeight,
			Width:  m.PageWidth,
		}

		// Handle the page resize event
		m.accountsPage, cmd = m.accountsPage.HandleMessage(pageMsg)
		cmds = append(cmds, cmd)

		m.keysPage, cmd = m.keysPage.HandleMessage(pageMsg)
		cmds = append(cmds, cmd)

		// Avoid triggering commands again
		return m, tea.Batch(cmds...)
	}

	// Ignore commands while open
	if !m.modal.Open {
		// Get Page Updates
		switch m.page {
		case app.AccountsPage:
			m.accountsPage, cmd = m.accountsPage.HandleMessage(msg)
		case app.KeysPage:
			m.keysPage, cmd = m.keysPage.HandleMessage(msg)
		}
		cmds = append(cmds, cmd)
	}

	// Run Modal Updates Last,
	// This ensures Page Behavior is checked before mutating modal state
	m.modal, cmd = m.modal.HandleMessage(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// View renders the viewport.Model
func (m ViewportViewModel) View() string {

	// Handle Page render
	var page tea.Model
	switch m.page {
	case app.AccountsPage:
		page = m.accountsPage
	case app.KeysPage:
		page = m.keysPage
	}

	if page == nil {
		return "Error loading page..."
	}

	m.modal.Parent = fmt.Sprintf("%s\n%s", m.headerView(), page.View())
	return m.modal.View()
}

// headerView generates the top elements
func (m ViewportViewModel) headerView() string {
	if m.TerminalHeight < 15 {
		return ""
	}

	if m.TerminalWidth < 90 {
		if m.protocol.View() == "" {
			return lipgloss.JoinVertical(lipgloss.Center, m.status.View())
		}
		return lipgloss.JoinVertical(lipgloss.Center, m.status.View(), m.protocol.View())
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, m.status.View(), m.protocol.View())
}

// NewViewportViewModel handles the construction of the TUI viewport
func NewViewportViewModel(state *internal.StateModel, client *api.ClientWithResponses) (*ViewportViewModel, error) {
	m := ViewportViewModel{
		Data: state,

		// Header
		status:   MakeStatusViewModel(state),
		protocol: MakeProtocolViewModel(state),

		// Pages
		accountsPage: accounts.New(state),
		keysPage:     keys.New("", state.ParticipationKeys),

		// Modal
		modal: modal.New("", false, state),

		// Current Page
		page: app.AccountsPage,
		// RPC client
		client: client,

		errorPage: exception.New(""),
	}

	return &m, nil
}
