package overlay

import (
	"github.com/algorandfoundation/hack-tui/api"
	"github.com/algorandfoundation/hack-tui/ui/modals"
	"github.com/algorandfoundation/hack-tui/ui/style"
	tea "github.com/charmbracelet/bubbletea"
)

type ShowModal *api.ParticipationKey

func EmitShowModal(key *api.ParticipationKey) tea.Cmd {
	return func() tea.Msg {
		return ShowModal(key)
	}
}

type ViewModel struct {
	Parent string
	Open   bool
	modal  *modal.ViewModel
}

func (m ViewModel) SetKey(key *api.ParticipationKey) {
	m.modal.SetKey(key)
}
func New(parent string, open bool, modal *modal.ViewModel) ViewModel {
	return ViewModel{
		Parent: parent,
		Open:   open,
		modal:  modal,
	}
}

func (m ViewModel) Init() tea.Cmd {
	return nil
}
func (m ViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.HandleMessage(msg)
}
func (m ViewModel) HandleMessage(msg tea.Msg) (ViewModel, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case modal.DeleteFinished:
		m.modal.Page = modal.InfoModal
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.modal.Page == modal.InfoModal {
				m.Open = false
			}
		}
	}
	m.modal, cmd = m.modal.HandleMessage(msg)
	return m, cmd
}
func (m ViewModel) View() string {
	if !m.Open {
		return m.Parent
	}
	return style.WithOverlay(m.modal.View(), m.Parent)
}
