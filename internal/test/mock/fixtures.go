package mock

import (
	"github.com/algorandfoundation/nodekit/api"
)

var VoteKey = []byte("TESTKEY")
var SelectionKey = []byte("TESTKEY")
var StateProofKey = []byte("TESTKEY")
var StateProofKeyTwo = []byte("TESTKEYTWO")
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
			StateProofKey:             &StateProofKeyTwo,
			VoteFirstValid:            0,
			VoteKeyDilution:           100,
			VoteLastValid:             30000,
			VoteParticipationKey:      nil,
		},
		LastBlockProposal: nil,
		LastStateProof:    nil,
		LastVote:          nil,
	},
	{
		Address:             "EXPIRED",
		EffectiveFirstValid: nil,
		EffectiveLastValid:  nil,
		Id:                  "12345",
		Key: api.AccountParticipation{
			SelectionParticipationKey: nil,
			StateProofKey:             &StateProofKeyTwo,
			VoteFirstValid:            0,
			VoteKeyDilution:           100,
			VoteLastValid:             1,
			VoteParticipationKey:      nil,
		},
		LastBlockProposal: nil,
		LastStateProof:    nil,
		LastVote:          nil,
	},
}
var abcEligibility = true

var abcParticipation = api.AccountParticipation{
	SelectionParticipationKey: SelectionKey,
	StateProofKey:             &StateProofKey,
	VoteFirstValid:            0,
	VoteKeyDilution:           100,
	VoteLastValid:             30000,
	VoteParticipationKey:      VoteKey,
}
var ABCAccount = api.Account{
	Address:                     "ABC",
	Amount:                      100000,
	AmountWithoutPendingRewards: 0,
	AppsLocalState:              nil,
	AppsTotalExtraPages:         nil,
	AppsTotalSchema:             nil,
	Assets:                      nil,
	AuthAddr:                    nil,
	CreatedApps:                 nil,
	CreatedAssets:               nil,
	IncentiveEligible:           &abcEligibility,
	LastHeartbeat:               nil,
	LastProposed:                nil,
	MinBalance:                  0,
	Participation:               &abcParticipation,
	PendingRewards:              0,
	RewardBase:                  nil,
	Rewards:                     0,
	Round:                       0,
	SigType:                     nil,
	Status:                      "Online",
	TotalAppsOptedIn:            0,
	TotalAssetsOptedIn:          0,
	TotalBoxBytes:               nil,
	TotalBoxes:                  nil,
	TotalCreatedApps:            0,
	TotalCreatedAssets:          0,
}
var ExpiredAccount = api.Account{
	Address:                     "EXPIRED",
	Amount:                      100000,
	AmountWithoutPendingRewards: 0,
	AppsLocalState:              nil,
	AppsTotalExtraPages:         nil,
	AppsTotalSchema:             nil,
	Assets:                      nil,
	AuthAddr:                    nil,
	CreatedApps:                 nil,
	CreatedAssets:               nil,
	IncentiveEligible:           &abcEligibility,
	LastHeartbeat:               nil,
	LastProposed:                nil,
	MinBalance:                  0,
	Participation:               &abcParticipation,
	PendingRewards:              0,
	RewardBase:                  nil,
	Rewards:                     0,
	Round:                       0,
	SigType:                     nil,
	Status:                      "Online",
	TotalAppsOptedIn:            0,
	TotalAssetsOptedIn:          0,
	TotalBoxBytes:               nil,
	TotalBoxes:                  nil,
	TotalCreatedApps:            0,
	TotalCreatedAssets:          0,
}
