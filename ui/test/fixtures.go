package test

import (
	"github.com/algorandfoundation/hack-tui/api"
	"github.com/algorandfoundation/hack-tui/internal"
	"github.com/algorandfoundation/hack-tui/ui/test/mock"
	"time"
)

var VoteKey = []byte("TESTKEY")
var SelectionKey = []byte("TESTKEY")
var StateProofKey = []byte("TESTKEY")
var Keys = []api.ParticipationKey{
	{
		Address:             "ABC",
		EffectiveFirstValid: nil,
		EffectiveLastValid:  nil,
		Id:                  "123",
		Key: api.AccountParticipation{
			SelectionParticipationKey: SelectionKey,
			StateProofKey:             &StateProofKey,
			VoteFirstValid:            0,
			VoteKeyDilution:           100,
			VoteLastValid:             30000,
			VoteParticipationKey:      VoteKey,
		},
		LastBlockProposal: nil,
		LastStateProof:    nil,
		LastVote:          nil,
	},
	{
		Address:             "ABC",
		EffectiveFirstValid: nil,
		EffectiveLastValid:  nil,
		Id:                  "1234",
		Key: api.AccountParticipation{
			SelectionParticipationKey: nil,
			StateProofKey:             nil,
			VoteFirstValid:            0,
			VoteKeyDilution:           100,
			VoteLastValid:             30000,
			VoteParticipationKey:      nil,
		},
		LastBlockProposal: nil,
		LastStateProof:    nil,
		LastVote:          nil,
	},
}

func GetState() *internal.StateModel {
	sm := &internal.StateModel{
		Status: internal.StatusModel{
			State:       internal.StableState,
			Version:     "v-test",
			Network:     "v-test-network",
			Voting:      false,
			NeedsUpdate: false,
			LastRound:   0,
		},
		Metrics: internal.MetricsModel{
			Enabled:   true,
			Window:    100,
			RoundTime: time.Second * 2,
			TPS:       2.5,
			RX:        0,
			TX:        0,
			LastTS:    time.Time{},
			LastRX:    0,
			LastTX:    0,
		},
		Accounts:          nil,
		ParticipationKeys: &Keys,
		Admin:             false,
		Watching:          false,
	}
	values := make(map[string]internal.Account)
	clock := new(mock.Clock)
	for _, key := range *sm.ParticipationKeys {
		val, ok := values[key.Address]
		if !ok {
			values[key.Address] = internal.Account{
				Address: key.Address,
				Status:  "Offline",
				Balance: 0,
				Expires: internal.GetExpiresTime(clock, key, sm),
				Keys:    1,
			}
		} else {
			val.Keys++
			values[key.Address] = val
		}
	}
	sm.Accounts = values

	return sm
}
