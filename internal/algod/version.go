package algod

import (
	"context"
	"errors"
	"fmt"
	"github.com/algorandfoundation/nodekit/api"
)

// VersionResponse represents information about the system version, including network, version, and channel details.
type VersionResponse struct {

	// Network is a string representing the identifier of the blockchain or network associated with the system.
	Network string

	// Version is a string representing the version of the system, typically formatted as major.minor.build-channel.
	Version string

	// Channel is a string representing the release channel of the system, such as stable, beta, or nightly.
	Channel string
}

// GetVersion retrieves system version information from the API client and processes it into a formatted VersionResponse.
func GetVersion(ctx context.Context, client api.ClientWithResponsesInterface) (VersionResponse, api.ResponseInterface, error) {
	var release VersionResponse
	v, err := client.GetVersionWithResponse(ctx)
	if v == nil {
		return release, v, errors.New(InvalidVersionResponseError)
	}
	if err != nil {
		return release, *v, err
	}
	if v.StatusCode() != 200 {
		return release, v, errors.New(InvalidVersionResponseError)
	}
	release.Version = fmt.Sprintf("v%d.%d.%d-%s",
		v.JSON200.Build.Major,
		v.JSON200.Build.Minor,
		v.JSON200.Build.BuildNumber,
		v.JSON200.Build.Channel,
	)
	release.Network = v.JSON200.GenesisId
	release.Channel = v.JSON200.Build.Channel

	return release, v, nil
}
