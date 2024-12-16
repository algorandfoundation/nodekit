package api

import (
	"errors"
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

type LatestCatchpointResponse struct {
	ResponseCode   int
	ResponseStatus string
	JSON200        string
}

func (r LatestCatchpointResponse) StatusCode() int {
	return r.ResponseCode
}
func (r LatestCatchpointResponse) Status() string {
	return r.ResponseStatus
}

func GetLatestCatchpointWithResponse(http HttpPkgInterface, network string) (LatestCatchpointResponse, error) {
	var response LatestCatchpointResponse

	var url CatchPointUrl
	switch network {
	case "fnet-v1", "fnet":
		url = FNet
	case "betanet-v1.0", "betanet":
		url = BetaNet
	case "testnet-v1.0", "testnet":
		url = TestNet
	case "mainnet-v1.0", "mainnet":
		url = MainNet
	}

	res, err := http.Get(string(url))
	if err != nil {
		return response, err
	}
	defer res.Body.Close()

	// Handle invalid codes as errors
	if res.StatusCode >= 300 {
		return response, errors.New(res.Status)
	}

	// Read from the response
	var body []byte
	body, err = io.ReadAll(res.Body)
	if err != nil {
		return response, err
	}
	// Set the body and return
	response.JSON200 = strings.Replace(string(body), "\n", "", -1)
	return response, nil
}
