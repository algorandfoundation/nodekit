package app

import (
	"context"
	"github.com/algorandfoundation/hack-tui/api"
	"github.com/algorandfoundation/hack-tui/internal"
	tea "github.com/charmbracelet/bubbletea"
)

type DeleteFinished struct {
	Err *error
	Id  string
}

type DeleteKey *api.ParticipationKey

func EmitDeleteKey(ctx context.Context, client *api.ClientWithResponses, id string) tea.Cmd {
	return func() tea.Msg {
		err := internal.DeletePartKey(ctx, client, id)
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

func GenerateCmd(account string, state *internal.StateModel) tea.Cmd {
	params := api.GenerateParticipationKeysParams{
		Dilution: nil,
		First:    int(state.Status.LastRound),
		Last:     int(state.Status.LastRound) + state.Offset,
	}

	key, err := internal.GenerateKeyPair(state.Context, state.Client, account, &params)
	if err != nil {
		return EmitModalEvent(ModalEvent{
			Key:     nil,
			Address: "",
			Err:     &err,
			Type:    ExceptionModal,
		})
	}

	return EmitModalEvent(ModalEvent{
		Key:     key,
		Address: key.Address,
		Err:     nil,
		Type:    InfoModal,
	})
}
