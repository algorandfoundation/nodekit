package modal

import (
	"context"
	"github.com/algorandfoundation/hack-tui/api"
	"github.com/algorandfoundation/hack-tui/internal"
	"github.com/algorandfoundation/hack-tui/ui/modals/confirm"
	"github.com/algorandfoundation/hack-tui/ui/modals/info"
	"github.com/algorandfoundation/hack-tui/ui/modals/transaction"
	"github.com/algorandfoundation/hack-tui/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func DeleteKeyCmd(ctx context.Context, client *api.ClientWithResponses, id string) tea.Cmd {
	return func() tea.Msg {
		err := internal.DeletePartKey(ctx, client, id)
		if err != nil {
			return DeleteFinished(err.Error())
		}
		return DeleteFinished(id)
	}
}

type DeleteFinished string

type DeleteKey *api.ParticipationKey

type Page string

const (
	InfoModal        Page = "accounts"
	ConfirmModal     Page = "confirm"
	TransactionModal Page = "transaction"
)

type ViewModel struct {
	// Width and Height
	Width  int
	Height int

	// State for Context/Client
	State *internal.StateModel

	// Views
	infoModal        *info.ViewModel
	transactionModal *transaction.ViewModel
	confirmModal     *confirm.ViewModel

	// Current Component Data
	title       string
	controls    string
	borderColor string
	Page        Page
}

func New(state *internal.StateModel) *ViewModel {
	return &ViewModel{
		Width:  0,
		Height: 0,

		State: state,

		infoModal:        info.New(state),
		transactionModal: transaction.New(state),
		confirmModal:     confirm.New(state),

		Page:        InfoModal,
		controls:    "",
		borderColor: "3",
	}
}
func (m ViewModel) SetKey(key *api.ParticipationKey) {
	m.infoModal.ActiveKey = key
	m.confirmModal.ActiveKey = key
	m.transactionModal.ActiveKey = key
}
func (m ViewModel) Init() tea.Cmd {
	return nil
}
func (m ViewModel) HandleMessage(msg tea.Msg) (*ViewModel, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case DeleteFinished:
		m.Page = InfoModal
	case confirm.Msg:
		if msg != nil {
			return &m, DeleteKeyCmd(m.State.Context, m.State.Client, msg.Id)
		} else {
			m.Page = InfoModal
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.Page = InfoModal
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
		return &m, tea.Batch(cmds...)
	}

	// Only trigger modal commands when they are active
	switch m.Page {
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
	}
	cmds = append(cmds, cmd)
	return &m, tea.Batch(cmds...)
}
func (m ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.HandleMessage(msg)
}

func (m ViewModel) View() string {
	var render = ""
	switch m.Page {
	case InfoModal:
		render = m.infoModal.View()
	case TransactionModal:
		render = m.transactionModal.View()
	case ConfirmModal:
		render = m.confirmModal.View()
	}
	width := lipgloss.Width(render) + 2
	height := lipgloss.Height(render)
	return style.WithNavigation(m.controls, style.WithTitle(m.title, style.ApplyBorder(width, height, m.borderColor).PaddingRight(1).PaddingLeft(1).Render(render)))
}
