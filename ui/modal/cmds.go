package modal

import (
	"github.com/algorandfoundation/hack-tui/api"
	tea "github.com/charmbracelet/bubbletea"
)

type Event struct {
	Key     *api.ParticipationKey
	Address string
	Type    string
}

type ShowModal Event

func EmitShowModal(evt Event) tea.Cmd {
	return func() tea.Msg {
		return ShowModal(evt)
	}
}

type DeleteFinished string

type DeleteKey *api.ParticipationKey
