package test

import (
	"context"
	"github.com/algorandfoundation/algorun-tui/api"
	"github.com/algorandfoundation/algorun-tui/internal"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	mock2 "github.com/algorandfoundation/algorun-tui/internal/test/mock"
	"time"
)

func GetState(client api.ClientWithResponsesInterface) *internal.StateModel {
	sm := &internal.StateModel{
		Status: algod.Status{
			State:       algod.StableState,
			Version:     "v-test",
			Network:     "v-test-network",
			Voting:      false,
			NeedsUpdate: false,
			LastRound:   0,

			Client:  client,
			HttpPkg: new(api.HttpPkg),
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
		ParticipationKeys: &mock2.Keys,
		Admin:             false,
		Watching:          false,
		Client:            client,
		Context:           context.Background(),
	}
	values := make(map[string]algod.Account)
	for _, key := range *sm.ParticipationKeys {
		val, ok := values[key.Address]
		if !ok {
			values[key.Address] = algod.Account{
				Address:           key.Address,
				Status:            "Offline",
				Balance:           0,
				IncentiveEligible: true,
				Expires:           nil,
				Keys:              1,
			}
		} else {
			val.Keys++
			values[key.Address] = val
		}
	}
	sm.Accounts = values

	return sm
}
