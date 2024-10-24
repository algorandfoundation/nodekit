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

func (s *StateModel) Watch(cb func(model *StateModel, err error), ctx context.Context, client *api.ClientWithResponses) {
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
		s.ParticipationKeys, err = GetPartKeys(ctx, client)
		if err != nil {
			cb(nil, err)
		}

		// Get Accounts
		s.Accounts = AccountsFromState(s, client)

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

		// Set Metrics
		s.Metrics.RoundTime = getAverageDuration(timings)
		s.Metrics.Window = len(timings)
		s.Metrics.TPS = getAverage(txns)

		// Trim data
		if len(timings) >= 100 {
			timings = timings[1:]
			txns = txns[1:]
		}

		lastRound = s.Status.LastRound
		cb(s, nil)
	}
}
