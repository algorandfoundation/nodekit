package app

import (
	"encoding/base64"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/internal/algod/participation"
	"strings"

	"github.com/algorandfoundation/nodekit/api"
	tea "github.com/charmbracelet/bubbletea"
)

func EmitCreateShortLink(offline bool, part *api.ParticipationKey, state *algod.StateModel) tea.Cmd {
	if part == nil || state == nil {
		return nil
	}

	var loraNetwork = strings.Replace(strings.Replace(state.Status.Network, "-v1.0", "", 1), "-v1", "", 1)
	if loraNetwork == "dockernet" || loraNetwork == "tuinet" {
		loraNetwork = "localnet"
	}

	if offline {
		res, err := participation.GetOfflineShortLink(state.HttpPkg, participation.OfflineShortLinkBody{
			Account: part.Address,
			Network: loraNetwork,
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

	res, err := participation.GetOnlineShortLink(state.HttpPkg, participation.OnlineShortLinkBody{
		Account:          part.Address,
		VoteKeyB64:       base64.RawURLEncoding.EncodeToString(part.Key.VoteParticipationKey),
		SelectionKeyB64:  base64.RawURLEncoding.EncodeToString(part.Key.SelectionParticipationKey),
		StateProofKeyB64: base64.RawURLEncoding.EncodeToString(*part.Key.StateProofKey),
		VoteFirstValid:   part.Key.VoteFirstValid,
		VoteLastValid:    part.Key.VoteLastValid,
		KeyDilution:      part.Key.VoteKeyDilution,
		Network:          loraNetwork,
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
