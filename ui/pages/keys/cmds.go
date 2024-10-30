package keys

import (
	"github.com/algorandfoundation/hack-tui/api"
	tea "github.com/charmbracelet/bubbletea"
)

type DeleteKey *api.ParticipationKey

type DeleteFinished string

func EmitDeleteKey(key *api.ParticipationKey) tea.Cmd {
	return func() tea.Msg {
		return DeleteKey(key)
	}
}

func EmitKeyDeleted() tea.Cmd {
	return func() tea.Msg {
		return DeleteFinished("Key deleted")
	}
}

// EmitKeySelected waits for and retrieves a new set of table rows from a given channel.
func EmitKeySelected(key *api.ParticipationKey) tea.Cmd {
	return func() tea.Msg {
		return key
	}
}
