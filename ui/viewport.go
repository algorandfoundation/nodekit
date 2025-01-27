package ui

import (
	"errors"
	"fmt"

	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/ui/app"
	"github.com/algorandfoundation/nodekit/ui/overlay"
	"github.com/algorandfoundation/nodekit/ui/pages/accounts"
	"github.com/algorandfoundation/nodekit/ui/pages/keys"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ViewportViewModel represents the state and view model for a viewport in the application.
type ViewportViewModel struct {
	PageWidth, PageHeight         int
	TerminalWidth, TerminalHeight int

	Data *algod.StateModel

	// Header Components
	status   StatusViewModel
	protocol ProtocolViewModel

	// Pages
	accountsPage accounts.ViewModel
	keysPage     keys.ViewModel

	modal overlay.ViewModel
	page  app.Page
}

// Init hooks for components
func (m ViewportViewModel) Init() tea.Cmd {
	return tea.Batch(
		m.modal.Init(),
		m.accountsPage.Init(),
		m.keysPage.Init(),
	)
}

// Update Handle the viewport lifecycle
func (m ViewportViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)
	// Handle Header and Modal Updates
	m.protocol, cmd = m.protocol.HandleMessage(msg)
	cmds = append(cmds, cmd)
	m.status, cmd = m.status.HandleMessage(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case *algod.StateModel:
		m.Data = msg
	// When a page message comes, set the current page
	case app.Page:
		m.page = msg
		return m, nil
	// When the Participation Key endpoint responds, check for keys remaining
	// and navigate back to accounts when te participation key list is empty.
	case app.DeleteFinished:
		if len(m.keysPage.Rows()) <= 1 {
			cmds = append(cmds, app.EmitShowPage(app.AccountsPage))
		}
	// Handle navigations between the different pages and modals
	case tea.KeyMsg:
		// When the modal is open, handle controls via the overlay component
		if m.modal.Open {
			m.modal, cmd = m.modal.HandleMessage(msg)
			cmds = append(cmds, cmd)
			return m, tea.Batch(cmds...)
		}

		// Otherwise let the viewport have focus on the inputs for the following global controls
		switch msg.String() {
		case "g":
			// Only open modal when it is closed and not syncing
			if m.Data.Status.State == algod.StableState && m.Data.Metrics.RoundTime > 0 {
				return m, tea.Sequence(
					app.EmitAccountSelected(m.accountsPage.SelectedAccount()),
					app.EmitShowModal(app.GenerateModal),
				)
			} else if m.Data.Status.State != algod.StableState || m.Data.Metrics.RoundTime == 0 {
				genErr := errors.New("Please wait until your node is fully synced")
				m.modal, cmd = m.modal.HandleMessage(genErr)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
		case "left":
			// No more pages to the left
			if m.page == app.AccountsPage {
				return m, nil
			}
			// Navigate to the Keys Page
			if m.page == app.KeysPage {
				return m, app.EmitShowPage(app.AccountsPage)
			}
		case "right":
			// No more pages to the right
			if m.page != app.AccountsPage {
				return m, nil
			}

			// Navigate to the keys page
			selAcc := m.accountsPage.SelectedAccount()
			if selAcc != nil {
				return m, tea.Sequence(app.EmitAccountSelected(selAcc), app.EmitShowPage(app.KeysPage))
			}

			// Nothing to do if there are no accounts
			return m, nil

		// Exit the application
		case "q", "ctrl+c":
			return m, tea.Quit

		}

		// Pass commands to the pages, depending on which is active
		if m.page == app.AccountsPage {
			m.accountsPage, cmd = m.accountsPage.HandleMessage(msg)
			cmds = append(cmds, cmd)
		}
		if m.page == app.KeysPage {
			m.keysPage, cmd = m.keysPage.HandleMessage(msg)
			cmds = append(cmds, cmd)
		}

		return m, tea.Batch(cmds...)

	// Override the page height for the page renders
	case tea.WindowSizeMsg:
		// Handle modal height
		m.modal, cmd = m.modal.HandleMessage(msg)
		cmds = append(cmds, cmd)

		m.TerminalWidth = msg.Width
		m.TerminalHeight = msg.Height
		m.PageWidth = msg.Width
		m.PageHeight = max(0, msg.Height-lipgloss.Height(m.headerView()))

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

	// Handle all other events
	m.accountsPage, cmd = m.accountsPage.HandleMessage(msg)
	cmds = append(cmds, cmd)
	m.keysPage, cmd = m.keysPage.HandleMessage(msg)
	cmds = append(cmds, cmd)
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
func NewViewportViewModel(state *algod.StateModel) (*ViewportViewModel, error) {
	m := ViewportViewModel{
		Data: state,

		// Header
		status:   MakeStatusViewModel(state),
		protocol: MakeProtocolViewModel(state),

		// Pages
		accountsPage: accounts.New(state),
		keysPage:     keys.New("", state.ParticipationKeys),

		// Modal
		modal: overlay.New("", false, state),
		// Current Page
		page: app.AccountsPage,
	}

	return &m, nil
}
