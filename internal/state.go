package internal

import (
	"context"
	"errors"
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
		if err != nil {
			cb(nil, err)
		}
		if status.StatusCode() != 200 {
			cb(nil, errors.New(status.Status()))
		}

		// Update Status
		s.Status.LastRound = uint64(status.JSON200.LastRound)

		// Fetch Keys
		s.UpdateKeys(ctx, client)

		// Run Round Averages and RX/TX every 5 rounds
		if s.Status.LastRound%5 == 0 {
			bm, err := GetBlockMetrics(ctx, client, s.Status.LastRound, s.Metrics.Window)
			if err != nil {
				cb(nil, err)
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
		s.Metrics.TX = res["algod_network_sent_bytes_total"]
		s.Metrics.RX = res["algod_network_received_bytes_total"]
	}
}
func (s *StateModel) UpdateAccounts() {
	s.Accounts = AccountsFromState(s)
}

func (s *StateModel) UpdateKeys(ctx context.Context, client *api.ClientWithResponses) {
	var err error
	s.ParticipationKeys, err = GetPartKeys(ctx, client)
	if err != nil {
		s.Admin = false
	}
	if err == nil {
		s.Admin = true
		s.UpdateAccounts()
	}
}
