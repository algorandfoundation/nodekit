package mock

import (
	"encoding/base64"
	"github.com/algorandfoundation/nodekit/api"
)

func GetPartKey() (*api.ParticipationKey, error) {
	account := "TUIDKH2C7MUHZDD77MAMUREJRKNK25SYXB7OAFA6JFBB24PEL5UX4S4GUU"
	selectionKey, err := base64.StdEncoding.DecodeString("DM9cyZ0oLuVHDtVzhkhLIW06uE0J9faf6aL/FeLFj3o=")
	if err != nil {
		return nil, err
	}
	stateProofKey, err := base64.StdEncoding.DecodeString("+DAZBTXOletJxFUhEaYQaWaNs3Q4DLEwOlJ68gI8IGq9Ss/1szOimQiAt+f6lqk4FxEe/XvaAXkMbv2/9OiE1g==")
	if err != nil {
		return nil, err
	}

	voteKey, err := base64.StdEncoding.DecodeString("dHynahCuNWpeR9BcE+B8VE1GM/KdUj759k9ja8zNY30=")
	if err != nil {
		return nil, err
	}

	var fv = 47733256
	var lv = 47861731
	var kd = 359
	return &api.ParticipationKey{
		Address:             account,
		EffectiveFirstValid: nil,
		EffectiveLastValid:  nil,
		Id:                  "DRINJEQ6PN4GDQYE6ECJFRTVSCXZA4BUWNZMPFL6Q2MFTEFBPXGA",
		Key: api.AccountParticipation{
			SelectionParticipationKey: selectionKey,
			StateProofKey:             &stateProofKey,
			VoteFirstValid:            fv,
			VoteKeyDilution:           kd,
			VoteLastValid:             lv,
			VoteParticipationKey:      voteKey,
		},
		LastBlockProposal: nil,
		LastStateProof:    nil,
		LastVote:          nil,
	}, nil
}

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
