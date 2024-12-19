package algod

import (
	"context"
	"errors"
	"github.com/algorandfoundation/algorun-tui/api"
)

// StartCatchup sends a request to start a catchup operation on a specific catchpoint and returns the catchup message.
// It uses the provided API client, catchpoint string, and optional parameters for catchup configuration.
// Returns the catchup message, the raw API response, and an error if any occurred.
func StartCatchup(ctx context.Context, client api.ClientWithResponsesInterface, catchpoint string, params *api.StartCatchupParams) (string, api.ResponseInterface, error) {
	response, err := client.StartCatchupWithResponse(ctx, catchpoint, params)
	if err != nil {
		return "", response, err
	}
	if response.StatusCode() != 200 {
		return "", response, errors.New(response.Status())
	}

	return response.JSON200.CatchupMessage, response, nil
}

// AbortCatchup aborts a ledger catchup process for the specified catchpoint using the provided client interface.
func AbortCatchup(ctx context.Context, client api.ClientWithResponsesInterface, catchpoint string) (string, api.ResponseInterface, error) {
	response, err := client.AbortCatchupWithResponse(ctx, catchpoint)
	if err != nil {
		return "", response, err
	}
	if response.StatusCode() != 200 {
		return "", response, errors.New(response.Status())
	}

	return response.JSON200.CatchupMessage, response, nil
}

// GetLatestCatchpoint fetches the latest catchpoint for the specified network using the provided HTTP package.
func GetLatestCatchpoint(httpPkg api.HttpPkgInterface, network string) (string, api.ResponseInterface, error) {
	response, err := api.GetLatestCatchpointWithResponse(httpPkg, network)
	if err != nil {
		return "", response, err
	}
	return response.JSON200, response, nil
}
