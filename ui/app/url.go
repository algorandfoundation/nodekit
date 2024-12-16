package app

import (
	"encoding/base64"
	"github.com/algorandfoundation/algorun-tui/api"
	"github.com/algorandfoundation/algorun-tui/internal"
	tea "github.com/charmbracelet/bubbletea"
)

func EmitCreateShortLink(offline bool, part *api.ParticipationKey, state *internal.StateModel) tea.Cmd {
	if part == nil || state == nil {
		return nil
	}
	if offline {
		res, err := internal.GetOfflineShortLink(state.Http, internal.OfflineShortLinkBody{
			Account: part.Address,
			Network: state.Status.Network,
		})
		if err != nil {
			return func() tea.Msg {
				return err
			}
		}
		return func() tea.Msg {
			return res
		}
	}

	res, err := internal.GetOnlineShortLink(state.Http, internal.OnlineShortLinkBody{
		Account:          part.Address,
		VoteKeyB64:       base64.RawURLEncoding.EncodeToString(part.Key.VoteParticipationKey),
		SelectionKeyB64:  base64.RawURLEncoding.EncodeToString(part.Key.SelectionParticipationKey),
		StateProofKeyB64: base64.RawURLEncoding.EncodeToString(*part.Key.StateProofKey),
		VoteFirstValid:   part.Key.VoteFirstValid,
		VoteLastValid:    part.Key.VoteLastValid,
		KeyDilution:      part.Key.VoteKeyDilution,
		Network:          state.Status.Network,
	})
	if err != nil {
		return func() tea.Msg {
			return err
		}
	}
	return func() tea.Msg {
		return res
	}
}
