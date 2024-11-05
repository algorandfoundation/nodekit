package internal

import (
	"context"
	"errors"
	"time"

	"github.com/algorandfoundation/hack-tui/api"
)

type StateModel struct {
	Status            StatusModel
	Metrics           MetricsModel
	Accounts          map[string]Account
	ParticipationKeys *[]api.ParticipationKey
	// TODO: handle contexts instead of adding it to state
	Admin    bool
	Watching bool
}

func (s *StateModel) waitAfterError(err error, cb func(model *StateModel, err error)) {
	if err != nil {
		s.Status.State = "DOWN"
		cb(nil, err)
		time.Sleep(time.Second * 3)
	}
}

// TODO: allow context to handle loop
func (s *StateModel) Watch(cb func(model *StateModel, err error), ctx context.Context, client *api.ClientWithResponses) {
	s.Watching = true
	if s.Metrics.Window == 0 {
		s.Metrics.Window = 100
	}

	err := s.Status.Fetch(ctx, client)
	if err != nil {
		cb(nil, err)
	}

	lastRound := s.Status.LastRound

	for {
		if !s.Watching {
			break
		}
		status, err := client.WaitForBlockWithResponse(ctx, int(lastRound))
		s.waitAfterError(err, cb)
		if err != nil {
			continue
		}
		if status.StatusCode() != 200 {
			s.waitAfterError(errors.New(status.Status()), cb)
			continue
		}

		s.Status.State = "Unknown"

		// Update Status
		s.Status.Update(status.JSON200.LastRound, status.JSON200.CatchupTime, status.JSON200.UpgradeNodeVote)

		// Fetch Keys
		s.UpdateKeys(ctx, client)

		// Run Round Averages and RX/TX every 5 rounds
		if s.Status.LastRound%5 == 0 {
			bm, err := GetBlockMetrics(ctx, client, s.Status.LastRound, s.Metrics.Window)
			s.waitAfterError(err, cb)
			if err != nil {
				continue
			}
			s.Metrics.RoundTime = bm.AvgTime
			s.Metrics.TPS = bm.TPS
			s.UpdateMetricsFromRPC(ctx, client)
		}

		lastRound = s.Status.LastRound
		cb(s, nil)
	}
}

func (s *StateModel) Stop() {
	s.Watching = false
}

func (s *StateModel) UpdateMetricsFromRPC(ctx context.Context, client *api.ClientWithResponses) {
	// Fetch RX/TX
	res, err := GetMetrics(ctx, client)
	if err != nil {
		s.Metrics.Enabled = false
	}
	if err == nil {
		s.Metrics.Enabled = true
		now := time.Now()
		diff := now.Sub(s.Metrics.LastTS)

		s.Metrics.TX = max(0, int(float64(res["algod_network_sent_bytes_total"]-s.Metrics.LastTX)/diff.Seconds()))
		s.Metrics.RX = max(0, int(float64(res["algod_network_received_bytes_total"]-s.Metrics.LastRX)/diff.Seconds()))

		s.Metrics.LastTS = now
		s.Metrics.LastTX = res["algod_network_sent_bytes_total"]
		s.Metrics.LastRX = res["algod_network_received_bytes_total"]
	}
}
func (s *StateModel) UpdateAccounts(client *api.ClientWithResponses) {
	s.Accounts = AccountsFromState(s, new(Clock), client)
}

func (s *StateModel) UpdateKeys(ctx context.Context, client *api.ClientWithResponses) {
	var err error
	s.ParticipationKeys, err = GetPartKeys(ctx, client)
	if err != nil {
		s.Admin = false
	}
	if err == nil {
		s.Admin = true
		s.UpdateAccounts(client)
	}
}
