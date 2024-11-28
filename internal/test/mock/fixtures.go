package mock

import (
	"github.com/algorandfoundation/hack-tui/api"
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
