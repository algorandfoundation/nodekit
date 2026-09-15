package statewalker

import (
	"context"
	"time"

	"github.com/algorandfoundation/nodekit/api"
	"github.com/algorandfoundation/nodekit/internal/algod"
)

// FallbackRoundTime is used when the chain is too young to measure an average
// block time from block timestamps.
const FallbackRoundTime = 2800 * time.Millisecond

// maxRoundTimeWindow bounds how many rounds are sampled when measuring the
// average block time.
const maxRoundTimeWindow = 20

// MeasureRoundTime returns the average block time over recent rounds using the
// same block-timestamp calculation as the nodekit TUI, falling back to
// FallbackRoundTime when not enough rounds exist yet.
func MeasureRoundTime(ctx context.Context, client api.ClientWithResponsesInterface) (time.Duration, error) {
	status, _, err := algod.Status{Client: client}.Get(ctx)
	if err != nil {
		return 0, err
	}
	window := maxRoundTimeWindow
	if int(status.LastRound)-1 < window {
		window = int(status.LastRound) - 1
	}
	if window < 1 {
		return FallbackRoundTime, nil
	}
	metrics, _, err := algod.GetBlockMetrics(ctx, client, status.LastRound, window)
	if err != nil {
		return 0, err
	}
	if metrics.AvgTime <= 0 {
		return FallbackRoundTime, nil
	}
	return metrics.AvgTime, nil
}

// DurationToRounds converts a wall-clock duration into a round count given an
// average round time.
func DurationToRounds(d time.Duration, roundTime time.Duration) uint64 {
	if roundTime <= 0 {
		roundTime = FallbackRoundTime
	}
	rounds := uint64(d / roundTime)
	if rounds < 1 {
		rounds = 1
	}
	return rounds
}
