package internal

import (
	"context"
	"errors"
	"github.com/algorandfoundation/hack-tui/api"
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
		return nil, errors.New("invalid status code")
	}

	return block.JSON200.Block, nil
}
