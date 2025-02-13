package app

import (
	"errors"
	"github.com/algorandfoundation/nodekit/api"
	"github.com/algorandfoundation/nodekit/internal/algod"
	tea "github.com/charmbracelet/bubbletea"
)

type FastCatchupStarted string
type FastCatchupStopped string

func StartFastCatchupCmd(state *algod.StateModel) tea.Cmd {
	return func() tea.Msg {
		threshold := algod.CATCHPOINT_THRESHOLD
		// Fetch catchpoint
		catchpoint, _, err := algod.GetLatestCatchpoint(state.HttpPkg, state.Status.Network)
		if err != nil {
			return err
		}
		if catchpoint == "" {
			return errors.New(algod.NO_CATCHPOINT)
		}
		// Submit the Catchpoint to the Algod Node, using the StartCatchupParams to skip
		res, _, err := algod.StartCatchup(state.Context, state.Client, catchpoint, &api.StartCatchupParams{Min: &threshold})
		if err != nil {
			return err
		}

		return FastCatchupStarted(res)
	}
}

// AbortFastCatchupCmd stops the fast catchup process if the system is in the FastCatchupState and returns the resulting message or error.
func AbortFastCatchupCmd(state *algod.StateModel) tea.Cmd {
	if state == nil {
		return nil
	}
	return func() tea.Msg {
		if state.Status.State == algod.FastCatchupState {
			res, _, err := algod.AbortCatchup(state.Context, state.Client, *state.Status.Catchpoint)
			if err != nil {
				return err
			}
			return FastCatchupStopped(res)
		}
		return nil
	}
}
