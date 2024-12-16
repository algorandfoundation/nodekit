package algod

import (
	"context"
	"errors"
	"github.com/algorandfoundation/algorun-tui/api"
)

func PostCatchpoint(ctx context.Context, client api.ClientWithResponsesInterface, catchpoint string, params *api.StartCatchupParams) (string, api.ResponseInterface, error) {
	response, err := client.StartCatchupWithResponse(ctx, catchpoint, params)
	if err != nil {
		return "", response, err
	}
	if response.StatusCode() != 200 {
		return "", response, errors.New(response.Status())
	}

	return response.JSON200.CatchupMessage, response, nil
}

func GetLatestCatchpoint(httpPkg api.HttpPkgInterface, network string) (string, api.ResponseInterface, error) {
	response, err := api.GetLatestCatchpointWithResponse(httpPkg, network)
	if err != nil {
		return "", response, err
	}
	return response.JSON200, response, nil
}
