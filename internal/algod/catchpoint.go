package algod

import (
	"context"
	"errors"
	"github.com/algorandfoundation/nodekit/api"
	"strconv"
	"strings"
)

const CATCHPOINT_THRESHOLD = 30_000
const NO_CATCHPOINT = "no catchpoint found"

// StartCatchup sends a request to start a catchup operation on a specific catchpoint and returns the catchup message.
// It uses the provided API client, catchpoint string, and optional parameters for catchup configuration.
// Returns the catchup message, the raw API response, and an error if any occurred.
func StartCatchup(ctx context.Context, client api.ClientWithResponsesInterface, catchpoint string, params *api.StartCatchupParams) (string, api.ResponseInterface, error) {
	response, err := client.StartCatchupWithResponse(ctx, catchpoint, params)
	if err != nil {
		return "", response, err
	}
	if response.StatusCode() >= 300 {
		return "", response, errors.New(response.Status())
	}
	if response.StatusCode() == 200 {
		return response.JSON200.CatchupMessage, response, nil
	}

	return response.JSON201.CatchupMessage, response, nil
}

// AbortCatchup aborts a ledger catchup process for the specified catchpoint using the provided client interface.
func AbortCatchup(ctx context.Context, client api.ClientWithResponsesInterface, catchpoint string) (string, api.ResponseInterface, error) {
	response, err := client.AbortCatchupWithResponse(ctx, catchpoint)
	if err != nil {
		return "", response, err
	}
	if response.StatusCode() >= 300 {
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

// IsLagging determines if the given round is lagging behind the network's latest catchpoint round by a predefined threshold.
// It takes an HTTP package interface, the current round, and the network name as inputs, and returns a boolean and an error.
func IsLagging(httpPkg api.HttpPkgInterface, round uint64, network string) (bool, error) {
	// Fetch catchpoint
	catchpoint, _, err := GetLatestCatchpoint(httpPkg, network)
	if err != nil {
		return false, err
	}
	if catchpoint == "" {
		return false, errors.New(NO_CATCHPOINT)
	}
	// Parse catchpoint round
	// Example: 48670000#AXHC4X4SSLE7QUSXE5CLPRPV2YUNK3EL6CFVEYWXGONNRO6GWXRQ
	catchpointParts := strings.Split(catchpoint, "#")
	catchpointRound, err := strconv.ParseUint(catchpointParts[0], 10, 64)
	if err != nil {
		return false, err
	}

	// Considered lagging if the delta of the rounds are above the threshold
	delta := int(catchpointRound) - int(round)
	return CATCHPOINT_THRESHOLD < delta, nil
}
