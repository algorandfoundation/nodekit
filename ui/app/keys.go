package app

import (
	"context"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/internal/algod/participation"
	"github.com/charmbracelet/lipgloss"
	"time"

	"github.com/algorandfoundation/nodekit/api"
	tea "github.com/charmbracelet/bubbletea"
)

// DeleteFinished represents the result of a deletion operation, containing an optional error and the associated ID.
type DeleteFinished struct {
	Err *error
	Id  string
}

// EmitDeleteKey creates a command to delete a participation key by ID and returns the result as a DeleteFinished message.
func EmitDeleteKey(ctx context.Context, client api.ClientWithResponsesInterface, id string) tea.Cmd {
	return func() tea.Msg {
		err := participation.Delete(ctx, client, id)
		if err != nil {
			return DeleteFinished{
				Err: &err,
				Id:  "",
			}
		}
		return DeleteFinished{
			Err: nil,
			Id:  id,
		}
	}
}

// GenerateCmd creates a command to generate participation keys for a specified account using given range type and duration.
// It utilizes the current state to configure the parameters required for key generation and returns a ModalEvent as a message.
func GenerateCmd(account string, rangeType participation.RangeType, duration int, state *algod.StateModel) tea.Cmd {
	return func() tea.Msg {
		var params api.GenerateParticipationKeysParams

		if rangeType == participation.TimeRange {
			params = api.GenerateParticipationKeysParams{
				Dilution: nil,
				First:    int(state.Status.LastRound),
				Last:     int(state.Status.LastRound) + int((time.Duration(duration) / state.Metrics.RoundTime)),
			}
		} else {
			params = api.GenerateParticipationKeysParams{
				Dilution: nil,
				First:    int(state.Status.LastRound),
				Last:     int(state.Status.LastRound) + int(duration),
			}
		}

		key, err := participation.GenerateKeys(state.Context, state.Client, account, &params)
		if err != nil {
			return err
		}

		return KeySelectedEvent{
			Key: key,
			Prefix: lipgloss.JoinVertical(
				lipgloss.Left,
				"Participation keys generated.",
				"",
				"Next step: register the participation keys with the network by signing a keyreg online transaction.",
				"Press the R key to start this process.",
				"",
			),
			Active: false,
		}
	}

}

// KeySelectedEvent represents an event triggered in the modal system.
type KeySelectedEvent struct {

	// Key represents a participation key associated with the modal event.
	Key *api.ParticipationKey

	// Active indicates whether key is Online or not.
	Active bool

	// Prefix adds prefix message to info modal
	Prefix string
}

// EmitKeySelectedEvent creates a command that emits a ModalEvent as a message in the Tea framework.
func EmitKeySelectedEvent(event KeySelectedEvent) tea.Cmd {
	return func() tea.Msg {
		return event
	}
}
