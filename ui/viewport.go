package ui

import (
	"context"
	"fmt"

	"github.com/algorandfoundation/hack-tui/api"
	"github.com/algorandfoundation/hack-tui/internal"
	"github.com/algorandfoundation/hack-tui/ui/pages/accounts"
	"github.com/algorandfoundation/hack-tui/ui/pages/generate"
	"github.com/algorandfoundation/hack-tui/ui/pages/keys"
	"github.com/algorandfoundation/hack-tui/ui/pages/transaction"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ViewportPage string

const (
	AccountsPage    ViewportPage = "accounts"
	KeysPage        ViewportPage = "keys"
	GeneratePage    ViewportPage = "generate"
	TransactionPage ViewportPage = "transaction"
	ErrorPage       ViewportPage = "error"
)

type ViewportViewModel struct {
	PageWidth, PageHeight         int
	TerminalWidth, TerminalHeight int

	Data *internal.StateModel

	// Header Components
	status   StatusViewModel
	protocol ProtocolViewModel

	// Pages
	accountsPage    accounts.ViewModel
	keysPage        keys.ViewModel
	generatePage    generate.ViewModel
	transactionPage transaction.ViewModel

	page   ViewportPage
	client *api.ClientWithResponses

	// Error Handler
	errorMsg  *string
	errorPage ErrorViewModel
}

func DeleteKey(client *api.ClientWithResponses, key keys.DeleteKey) tea.Cmd {
	return func() tea.Msg {
		err := internal.DeletePartKey(context.Background(), client, key.Id)
		if err != nil {
			return keys.DeleteFinished(err.Error())
		}
		return keys.DeleteFinished("Key deleted")
	}
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
	case error:
		strMsg := msg.Error()
		m.errorMsg = &strMsg
	// When the state updates
	case internal.StateModel:
		if m.errorMsg != nil {
			m.errorMsg = nil
			m.page = AccountsPage
		}
		m.Data = &msg
		// Navigate to the transaction page when a partkey is selected
	case *api.ParticipationKey:
		m.page = TransactionPage
	// Navigate to the keys page when an account is selected
	case internal.Account:
		m.page = KeysPage
	case keys.DeleteKey:
		return m, DeleteKey(m.client, msg)
	case keys.DeleteFinished:
		return m, keys.EmitKeyDeleted()
	case tea.KeyMsg:
		switch msg.String() {
		// Tab Backwards
		case "shift+tab":
			if m.page == AccountsPage {
				return m, nil
			}
			if m.page == TransactionPage {
				return m, accounts.EmitAccountSelected(m.accountsPage.SelectedAccount())
			}
			if m.page == KeysPage {
				m.page = AccountsPage
				return m, nil
			}
		// Tab Forwards
		case "tab":
			if m.page == AccountsPage {
				m.page = KeysPage
				return m, accounts.EmitAccountSelected(m.accountsPage.SelectedAccount())
			}
			if m.page == KeysPage {
				m.page = TransactionPage
				// If there isn't a key already, select the first record
				if m.keysPage.SelectedKey() == nil && m.Data != nil {
					data := *m.Data.ParticipationKeys
					return m, keys.EmitKeySelected(&data[0])
				}
				// Navigate to the transaction page
				return m, keys.EmitKeySelected(m.keysPage.SelectedKey())
			}
		case "a":
			m.page = AccountsPage
		case "g":
			m.generatePage.Inputs[0].SetValue(m.accountsPage.SelectedAccount().Address)
			m.page = GeneratePage
			return m, nil
		case "k":
			m.page = KeysPage
			return m, accounts.EmitAccountSelected(m.accountsPage.SelectedAccount())
		case "t":
			m.page = TransactionPage
			// If there isn't a key already, select the first record for that account
			if m.keysPage.SelectedKey() == nil && m.Data != nil {
				data := *m.Data.ParticipationKeys
				acct := m.accountsPage.SelectedAccount()
				for i, key := range data {
					if key.Address == acct.Address {
						return m, keys.EmitKeySelected(&data[i])
					}
				}
			}
			// Navigate to the transaction page
			return m, keys.EmitKeySelected(m.keysPage.SelectedKey())
		case "ctrl+c":
			return m, tea.Quit
		}

	case tea.WindowSizeMsg:
		m.TerminalWidth = msg.Width
		m.TerminalHeight = msg.Height
		m.PageWidth = msg.Width
		m.PageHeight = max(0, msg.Height-lipgloss.Height(m.headerView())-1)

		// Custom size message
		pageMsg := tea.WindowSizeMsg{
			Height: m.PageHeight,
			Width:  m.PageWidth,
		}

		// Handle the page resize event
		//switch m.page {
		//case AccountsPage:
		m.accountsPage, cmd = m.accountsPage.HandleMessage(pageMsg)
		cmds = append(cmds, cmd)
		//case KeysPage:
		m.keysPage, cmd = m.keysPage.HandleMessage(pageMsg)
		cmds = append(cmds, cmd)
		//case GeneratePage:
		m.generatePage, cmd = m.generatePage.HandleMessage(pageMsg)
		cmds = append(cmds, cmd)
		//case TransactionPage:
		m.transactionPage, cmd = m.transactionPage.HandleMessage(pageMsg)
		cmds = append(cmds, cmd)
		//}
		m.errorPage, cmd = m.errorPage.HandleMessage(pageMsg)
		cmds = append(cmds, cmd)
		// Avoid triggering commands again
		return m, tea.Batch(cmds...)

	}
	// Get Page Updates
	switch m.page {
	case AccountsPage:
		m.accountsPage, cmd = m.accountsPage.HandleMessage(msg)
	case KeysPage:
		m.keysPage, cmd = m.keysPage.HandleMessage(msg)
	case GeneratePage:
		m.generatePage, cmd = m.generatePage.HandleMessage(msg)
	case TransactionPage:
		m.transactionPage, cmd = m.transactionPage.HandleMessage(msg)
	}
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// View renders the viewport.Model
func (m ViewportViewModel) View() string {
	errMsg := m.errorMsg

	if errMsg != nil {
		m.errorPage.Message = *errMsg
		m.page = ErrorPage
	}

	// Handle Page render
	var page tea.Model
	switch m.page {
	case AccountsPage:
		page = m.accountsPage
	case GeneratePage:
		page = m.generatePage
	case KeysPage:
		page = m.keysPage
	case TransactionPage:
		page = m.transactionPage
	case ErrorPage:
		page = m.errorPage
	}

	if page == nil {
		return "Error loading page..."
	}

	return fmt.Sprintf("%s\n%s", m.headerView(), page.View())
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

// MakeViewportViewModel handles the construction of the TUI viewport
func MakeViewportViewModel(state *internal.StateModel, client *api.ClientWithResponses) (*ViewportViewModel, error) {
	m := ViewportViewModel{
		Data: state,

		// Header
		status:   MakeStatusViewModel(state),
		protocol: MakeProtocolViewModel(state),

		// Pages
		accountsPage:    accounts.New(state),
		keysPage:        keys.New("", state.ParticipationKeys),
		generatePage:    generate.New("", client),
		transactionPage: transaction.New(state),

		// Current Page
		page: AccountsPage,
		// RPC client
		client: client,

		errorPage: NewErrorViewModel(""),
	}

	return &m, nil
}
