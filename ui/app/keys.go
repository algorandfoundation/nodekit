package app

import (
	"context"
	"github.com/algorandfoundation/algorun-tui/internal/algod/participation"
	"time"

	"github.com/algorandfoundation/algorun-tui/api"
	"github.com/algorandfoundation/algorun-tui/internal"
	tea "github.com/charmbracelet/bubbletea"
)

type DeleteFinished struct {
	Err *error
	Id  string
}

type DeleteKey *api.ParticipationKey

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

func GenerateCmd(account string, rangeType participation.RangeType, duration int, state *internal.StateModel) tea.Cmd {
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
			return ModalEvent{
				Key:     nil,
				Address: "",
				Active:  false,
				Err:     &err,
				Type:    ExceptionModal,
			}
		}

		return ModalEvent{
			Key:     key,
			Address: key.Address,
			Active:  false,
			Err:     nil,
			Type:    InfoModal,
		}
	}

}
