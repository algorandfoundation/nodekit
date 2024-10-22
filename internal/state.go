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

func getAverage(data []float64) float64 {
	sum := 0.0
	for _, element := range data {
		sum += element
	}
	return sum / (float64(len(data)))
}
func getAverageDuration(timings []time.Duration) time.Duration {
	sum := 0.0
	for _, element := range timings {
		sum += element.Seconds()
	}
	avg := sum / (float64(len(timings)))
	return time.Duration(avg * float64(time.Second))
}

// TODO: allow context to handle loop
func (s *StateModel) Watch(cb func(model *StateModel, err error), ctx context.Context, client *api.ClientWithResponses) {
	s.Watching = true

	err := s.Status.Fetch(ctx, client)
	if err != nil {
		cb(nil, err)
	}

	lastRound := s.Status.LastRound

	// Collection of Round Durations
	timings := make([]time.Duration, 0)
	// Collection of Transaction Counts
	txns := make([]float64, 0)

	for {
		if !s.Watching {
			break
		}
		// Collect Time of Round
		startTime := time.Now()
		status, err := client.WaitForBlockWithResponse(ctx, int(lastRound))
		if err != nil {
			cb(nil, err)
		}
		if status.StatusCode() != 200 {
			cb(nil, errors.New(status.Status()))
		}
		// Store round timing
		endTime := time.Now()
		dur := endTime.Sub(startTime)
		timings = append(timings, dur)

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

		// Check for transactions
		if block.JSON200.Block["txns"] != nil {
			// Get the average duration in seconds (TPS)
			txnCount := float64(len(block.JSON200.Block["txns"].([]any)))
			txns = append(txns, txnCount/getAverageDuration(timings).Seconds())
		} else {
			txns = append(txns, 0)
		}

		// Fetch RX/TX every 5th round
		if s.Status.LastRound%5 == 0 {
			s.UpdateMetrics(ctx, client, timings, txns)
		}
		// Trim data
		if len(timings) >= 100 {
			timings = timings[1:]
			txns = txns[1:]
		}

		lastRound = s.Status.LastRound
		cb(s, nil)
	}
}

func (s *StateModel) Stop() {
	s.Watching = false
}

func (s *StateModel) UpdateMetrics(
	ctx context.Context,
	client *api.ClientWithResponses,
	timings []time.Duration,
	txns []float64,
) {
	if s == nil {
		panic("StateModel is nil while UpdateMetrics is called")
	}
	// Set Metrics
	s.Metrics.RoundTime = getAverageDuration(timings)
	s.Metrics.Window = len(timings)
	s.Metrics.TPS = getAverage(txns)

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
