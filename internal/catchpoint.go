package internal

import (
	"context"
	"errors"
	"github.com/algorandfoundation/algorun-tui/api"
	"io"
	"strings"
)

type CatchPointUrl string

const (
	FNet    CatchPointUrl = "https://fnet-catchpoints.algorand.green/latest"
	BetaNet CatchPointUrl = "https://algorand-catchpoints.s3.us-east-2.amazonaws.com/channel/betanet/latest.catchpoint"
	TestNet CatchPointUrl = "https://algorand-catchpoints.s3.us-east-2.amazonaws.com/channel/testnet/latest.catchpoint"
	MainNet CatchPointUrl = "https://algorand-catchpoints.s3.us-east-2.amazonaws.com/channel/mainnet/latest.catchpoint"
)

func PostCatchpoint(ctx context.Context, client api.ClientWithResponsesInterface, catchpoint string, params *api.StartCatchupParams) (string, error) {
	res, err := client.StartCatchupWithResponse(ctx, catchpoint, params)
	if err != nil {
		return "", err
	}
	if res.StatusCode() != 200 {
		return "", errors.New(res.Status())
	}

	return res.JSON200.CatchupMessage, nil
}

func GetLatestCatchpoint(ctx context.Context, http HttpPkgInterface, network string) (string, error) {
	var catchpoint string
	var url CatchPointUrl
	switch network {
	case "fnet":
		url = FNet
	case "betanet":
		url = BetaNet
	case "testnet":
		url = TestNet
	case "mainnet":
		url = MainNet
	}

	res, err := http.Get(string(url))
	if err != nil {
		return catchpoint, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return catchpoint, errors.New(res.Status)
	}

	var body []byte
	body, err = io.ReadAll(res.Body)
	if err != nil {
		return catchpoint, err
	}

	return strings.Replace(string(body), "\n", "", -1), nil
}
