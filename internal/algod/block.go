package algod

import (
	"context"
	"errors"
	"github.com/algorandfoundation/nodekit/api"
	"time"
)

// InvalidWindow indicates the requested metrics window extends before the start of the chain.
const InvalidWindow = "invalid window"

// BlockMetrics represents metrics of a blockchain segment, including average block time and transactions per second.
type BlockMetrics struct {
	AvgTime time.Duration
	TPS     float64
}

// GetBlockMetrics calculates block metrics such as average block time and transactions per second within a specified window.
func GetBlockMetrics(ctx context.Context, client api.ClientWithResponsesInterface, round uint64, window int) (BlockMetrics, api.ResponseInterface, error) {
	var avgs = BlockMetrics{
		AvgTime: 0,
		TPS:     0,
	}
	var format api.GetBlockParamsFormat = "json"

	// Rounds are unsigned. A window wider than the current height would
	// underflow into an enormous round rather than an obviously-bogus one,
	// so reject it here instead of relying on every caller to pre-check.
	if window < 0 || round < uint64(window) {
		return avgs, nil, errors.New(InvalidWindow)
	}

	// Current Block
	currentBlockResponse, err := client.GetBlockWithResponse(ctx, round, &api.GetBlockParams{
		Format: &format,
	})
	if err != nil {
		return avgs, currentBlockResponse, err
	}
	if currentBlockResponse.StatusCode() != 200 {
		return avgs, currentBlockResponse, errors.New(currentBlockResponse.Status())
	}

	// Previous Block Response
	previousBlockResponse, err := client.GetBlockWithResponse(ctx, round-uint64(window), &api.GetBlockParams{
		Format: &format,
	})
	if err != nil {
		return avgs, previousBlockResponse, err
	}
	if previousBlockResponse.StatusCode() != 200 {
		return avgs, previousBlockResponse, errors.New(previousBlockResponse.Status())
	}

	// Push to the transactions count list
	aTimestampRes := currentBlockResponse.JSON200.Block["ts"]
	bTimestampRes := previousBlockResponse.JSON200.Block["ts"]
	if aTimestampRes == nil || bTimestampRes == nil {
		return avgs, previousBlockResponse, nil
	}
	aTimestamp := time.Duration(aTimestampRes.(float64)) * time.Second
	bTimestamp := time.Duration(bTimestampRes.(float64)) * time.Second

	// Transaction Counter
	aTransactions := currentBlockResponse.JSON200.Block["tc"]
	bTransactions := previousBlockResponse.JSON200.Block["tc"]

	avgs.AvgTime = time.Duration((int(aTimestamp - bTimestamp)) / window)
	if aTransactions != nil && bTransactions != nil {
		avgs.TPS = (aTransactions.(float64) - bTransactions.(float64)) / (float64(window) * avgs.AvgTime.Seconds())
	}

	return avgs, currentBlockResponse, nil
}
