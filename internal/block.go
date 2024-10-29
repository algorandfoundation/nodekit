package internal

import (
	"context"
	"errors"
	"github.com/algorandfoundation/hack-tui/api"
	"time"
)

func GetBlock(ctx context.Context, client *api.ClientWithResponses, round uint64) (map[string]interface{}, error) {

	var format api.GetBlockParamsFormat = "json"
	block, err := client.GetBlockWithResponse(ctx, int(round), &api.GetBlockParams{
		Format: &format,
	})
	if err != nil {
		return nil, err
	}

	if block.StatusCode() != 200 {
		return nil, errors.New(block.Status())
	}

	return block.JSON200.Block, nil
}

type BlockMetrics struct {
	AvgTime time.Duration
	TPS     float64
}

func GetBlockMetrics(ctx context.Context, client *api.ClientWithResponses, round uint64, window int) (*BlockMetrics, error) {
	var avgs = BlockMetrics{
		AvgTime: 0,
		TPS:     0,
	}
	if round < uint64(window) {
		return &avgs, nil
	}
	var format api.GetBlockParamsFormat = "json"
	a, err := client.GetBlockWithResponse(ctx, int(round), &api.GetBlockParams{
		Format: &format,
	})
	if err != nil {
		return nil, err
	}
	if a.StatusCode() != 200 {
		return nil, errors.New(a.Status())
	}
	b, err := client.GetBlockWithResponse(ctx, int(round)-window, &api.GetBlockParams{
		Format: &format,
	})
	if err != nil {
		return nil, err
	}
	if b.StatusCode() != 200 {
		return nil, errors.New(b.Status())
	}

	// Push to the transactions count list
	aTimestamp := time.Duration(a.JSON200.Block["ts"].(float64)) * time.Second
	bTimestamp := time.Duration(b.JSON200.Block["ts"].(float64)) * time.Second
	// Transaction Counter
	aTransactions := a.JSON200.Block["tc"]
	bTransactions := b.JSON200.Block["tc"]

	avgs.AvgTime = time.Duration((int(aTimestamp - bTimestamp)) / window)
	if aTransactions != nil && bTransactions != nil {
		avgs.TPS = (aTransactions.(float64) - bTransactions.(float64)) / (float64(window) * avgs.AvgTime.Seconds())
	}

	return &avgs, nil
}
