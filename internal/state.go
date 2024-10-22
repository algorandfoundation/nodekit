package internal

import (
	"context"
	"errors"
	"github.com/algorandfoundation/hack-tui/api"
	"time"
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

	// Collection of Transaction Counts
	txns := make([]float64, 0)

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

		// Fetch Block
		var format api.GetBlockParamsFormat = "json"
		block, err := client.GetBlockWithResponse(ctx, int(lastRound), &api.GetBlockParams{
			Format: &format,
		})
		if err != nil {
			cb(nil, err)
		}

		// Push to the transactions count list
		txnsValue := block.JSON200.Block["txns"]
		if txnsValue == nil {
			txns = append(txns, 0.0)
		}
		if txnsValue != nil {
			txns = append(txns, float64(len(txnsValue.([]interface{}))))
		}

		// Run Round Averages and RX/TX every 5 rounds
		if s.Status.LastRound%5 == 0 {
			s.UpdateMetricsFromRPC(ctx, client)
			err := s.UpdateRoundTime(ctx, client, time.Duration(block.JSON200.Block["ts"].(float64))*time.Second)
			if err != nil {
				cb(nil, err)
			}
		}
		txnSum := 0.0
		for i := 0; i < len(txns); i++ {
			txnSum += txns[i]
		}
		txnAvg := txnSum / float64(len(txns))
		if s.Metrics.RoundTime != 0 {
			s.Metrics.TPS = txnAvg / s.Metrics.RoundTime.Seconds()
		}

		// Trim data
		if len(txns) >= s.Metrics.Window {
			txns = txns[1:]
		}

		lastRound = s.Status.LastRound
		cb(s, nil)
	}
}

func (s *StateModel) Stop() {
	s.Watching = false
}

func (s *StateModel) UpdateRoundTime(
	ctx context.Context,
	client *api.ClientWithResponses,
	timestamp time.Duration,
) error {
	if s == nil {
		panic("StateModel is nil while UpdateMetrics is called")
	}
	previousRound := s.Status.LastRound - uint64(s.Metrics.Window)
	previousBlock, err := GetBlock(ctx, client, previousRound)
	if err != nil {
		s.Metrics.Enabled = false
		return err
	}
	previousBlockTs := time.Duration(previousBlock["ts"].(float64)) * time.Second

	s.Metrics.RoundTime = time.Duration(int(timestamp-previousBlockTs) / s.Metrics.Window)
	return nil
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
