package algod

import (
	"context"
	"errors"
	"fmt"
	"github.com/algorandfoundation/algorun-tui/api"
)

type VersionResponse struct {
	Network string
	Version string
	Channel string
}

func GetVersion(ctx context.Context, client api.ClientWithResponsesInterface) (VersionResponse, api.ResponseInterface, error) {
	var release VersionResponse
	v, err := client.GetVersionWithResponse(ctx)
	if err != nil {
		return release, v, err
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
