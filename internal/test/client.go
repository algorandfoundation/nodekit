package test

import (
	"context"
	"errors"
	"github.com/algorandfoundation/algorun-tui/api"
	"github.com/algorandfoundation/algorun-tui/internal/test/mock"
	"net/http"
)

func GetClient(throws bool) api.ClientWithResponsesInterface {
	return NewClient(throws, false)
}

type Client struct {
	api.ClientWithResponsesInterface
	Errors  bool
	Invalid bool
}

func NewClient(throws bool, invalid bool) api.ClientWithResponsesInterface {
	client := new(Client)
	if throws {
		client.Errors = true
	}
	if invalid {
		client.Invalid = true
	}
	return client
}

func (c *Client) MetricsWithResponse(ctx context.Context, reqEditors ...api.RequestEditorFn) (*api.MetricsResponse, error) {
	var res api.MetricsResponse
	body := `# HELP algod_telemetry_drops_total telemetry messages dropped due to full queues
# TYPE algod_telemetry_drops_total counter
algod_telemetry_drops_total 0
# HELP algod_telemetry_errs_total telemetry messages dropped due to server error
# TYPE algod_telemetry_errs_total counter
algod_telemetry_errs_total 0
# HELP algod_ram_usage number of bytes runtime.ReadMemStats().HeapInuse
# TYPE algod_ram_usage gauge
algod_ram_usage 0
# HELP algod_crypto_vrf_generate_total Total number of calls to GenerateVRFSecrets
# TYPE algod_crypto_vrf_generate_total counter
algod_crypto_vrf_generate_total 0
# HELP algod_crypto_vrf_prove_total Total number of calls to VRFSecrets.Prove
# TYPE algod_crypto_vrf_prove_total counter
algod_crypto_vrf_prove_total 0
# HELP algod_crypto_vrf_hash_total Total number of calls to VRFProof.Hash
# TYPE algod_crypto_vrf_hash_total counter
algod_crypto_vrf_hash_total 0`
	if !c.Invalid {
		httpResponse := http.Response{StatusCode: 200}
		res = api.MetricsResponse{
			Body:         []byte(body),
			HTTPResponse: &httpResponse,
		}
	} else {
		httpResponse := http.Response{StatusCode: 404}
		res = api.MetricsResponse{
			Body:         []byte(body),
			HTTPResponse: &httpResponse,
		}
	}
	if c.Errors {
		return &res, errors.New("test error")
	}
	return &res, nil
}
func (c *Client) GetParticipationKeyByIDWithResponse(ctx context.Context, participationId string, reqEditors ...api.RequestEditorFn) (*api.GetParticipationKeyByIDResponse, error) {
	var res api.GetParticipationKeyByIDResponse
	if !c.Invalid {
		httpResponse := http.Response{StatusCode: 200}
		res = api.GetParticipationKeyByIDResponse{
			Body:         nil,
			HTTPResponse: &httpResponse,
			JSON200:      &mock.Keys[0],
			JSON400:      nil,
			JSON401:      nil,
			JSON404:      nil,
			JSON500:      nil,
		}
	} else {
		httpResponse := http.Response{StatusCode: 404}
		res = api.GetParticipationKeyByIDResponse{
			Body:         nil,
			HTTPResponse: &httpResponse,
			JSON200:      nil,
			JSON400:      nil,
			JSON401:      nil,
			JSON404:      nil,
			JSON500:      nil,
		}
	}
	if c.Errors {
		return nil, errors.New("test error")
	}
	return &res, nil
}
func (c *Client) GetParticipationKeysWithResponse(ctx context.Context, reqEditors ...api.RequestEditorFn) (*api.GetParticipationKeysResponse, error) {
	var res api.GetParticipationKeysResponse
	clone := mock.Keys
	if !c.Invalid {
		httpResponse := http.Response{StatusCode: 200}
		res = api.GetParticipationKeysResponse{
			Body:         nil,
			HTTPResponse: &httpResponse,
			JSON200:      &clone,
			JSON400:      nil,
			JSON401:      nil,
			JSON404:      nil,
			JSON500:      nil,
		}
	} else {
		httpResponse := http.Response{StatusCode: 404}
		res = api.GetParticipationKeysResponse{
			Body:         nil,
			HTTPResponse: &httpResponse,
			JSON200:      &clone,
			JSON400:      nil,
			JSON401:      nil,
			JSON404:      nil,
			JSON500:      nil,
		}
	}

	if c.Errors {
		return nil, errors.New("test error")
	}
	return &res, nil
}

func (c *Client) DeleteParticipationKeyByIDWithResponse(ctx context.Context, participationId string, reqEditors ...api.RequestEditorFn) (*api.DeleteParticipationKeyByIDResponse, error) {
	var res api.DeleteParticipationKeyByIDResponse
	if !c.Invalid {
		httpResponse := http.Response{StatusCode: 200}
		res = api.DeleteParticipationKeyByIDResponse{
			Body:         nil,
			HTTPResponse: &httpResponse,
			JSON400:      nil,
			JSON401:      nil,
			JSON404:      nil,
			JSON500:      nil,
		}
	} else {
		httpResponse := http.Response{StatusCode: 404}
		res = api.DeleteParticipationKeyByIDResponse{
			Body:         nil,
			HTTPResponse: &httpResponse,
		}
	}

	if c.Errors {
		return &res, errors.New("test error")
	}
	return &res, nil
}

func (c *Client) AccountInformationWithResponse(ctx context.Context, address string, params *api.AccountInformationParams, reqEditors ...api.RequestEditorFn) (*api.AccountInformationResponse, error) {
	httpResponse := http.Response{StatusCode: 200}
	return &api.AccountInformationResponse{
		Body:         nil,
		HTTPResponse: &httpResponse,
		JSON200:      &mock.ABCAccount,
		JSON400:      nil,
		JSON401:      nil,
		JSON500:      nil,
	}, nil
}

func (c *Client) GenerateParticipationKeysWithResponse(ctx context.Context, address string, params *api.GenerateParticipationKeysParams, reqEditors ...api.RequestEditorFn) (*api.GenerateParticipationKeysResponse, error) {
	mock.Keys = append(mock.Keys, api.ParticipationKey{
		Address:             "ABC",
		EffectiveFirstValid: nil,
		EffectiveLastValid:  nil,
		Id:                  "",
		Key: api.AccountParticipation{
			SelectionParticipationKey: nil,
			StateProofKey:             nil,
			VoteFirstValid:            0,
			VoteKeyDilution:           0,
			VoteLastValid:             30,
			VoteParticipationKey:      nil,
		},
		LastBlockProposal: nil,
		LastStateProof:    nil,
		LastVote:          nil,
	})
	httpResponse := http.Response{StatusCode: 200}
	res := api.GenerateParticipationKeysResponse{
		Body:         nil,
		HTTPResponse: &httpResponse,
		JSON200:      nil,
		JSON400:      nil,
		JSON401:      nil,
		JSON500:      nil,
	}

	return &res, nil
}

func (c *Client) GetVersionWithResponse(ctx context.Context, reqEditors ...api.RequestEditorFn) (*api.GetVersionResponse, error) {
	var res api.GetVersionResponse
	version := api.Version{
		Build: api.BuildVersion{
			Branch:      "test",
			BuildNumber: 1,
			Channel:     "beta",
			CommitHash:  "abc",
			Major:       0,
			Minor:       0,
		},
		GenesisHashB64: nil,
		GenesisId:      "tui-net",
		Versions:       nil,
	}
	if !c.Invalid {
		httpResponse := http.Response{StatusCode: 200}

		res = api.GetVersionResponse{
			Body:         nil,
			HTTPResponse: &httpResponse,
			JSON200:      &version,
		}
	} else {
		httpResponse := http.Response{StatusCode: 404}
		res = api.GetVersionResponse{
			Body:         nil,
			HTTPResponse: &httpResponse,
			JSON200:      nil,
		}
	}
	if c.Errors {
		return &res, errors.New("test error")
	}
	return &res, nil
}
func (c *Client) GetStatusWithResponse(ctx context.Context, reqEditors ...api.RequestEditorFn) (*api.GetStatusResponse, error) {
	httpResponse := http.Response{StatusCode: 200}
	data := new(struct {
		Catchpoint                    *string `json:"catchpoint,omitempty"`
		CatchpointAcquiredBlocks      *int    `json:"catchpoint-acquired-blocks,omitempty"`
		CatchpointProcessedAccounts   *int    `json:"catchpoint-processed-accounts,omitempty"`
		CatchpointProcessedKvs        *int    `json:"catchpoint-processed-kvs,omitempty"`
		CatchpointTotalAccounts       *int    `json:"catchpoint-total-accounts,omitempty"`
		CatchpointTotalBlocks         *int    `json:"catchpoint-total-blocks,omitempty"`
		CatchpointTotalKvs            *int    `json:"catchpoint-total-kvs,omitempty"`
		CatchpointVerifiedAccounts    *int    `json:"catchpoint-verified-accounts,omitempty"`
		CatchpointVerifiedKvs         *int    `json:"catchpoint-verified-kvs,omitempty"`
		CatchupTime                   int     `json:"catchup-time"`
		LastCatchpoint                *string `json:"last-catchpoint,omitempty"`
		LastRound                     int     `json:"last-round"`
		LastVersion                   string  `json:"last-version"`
		NextVersion                   string  `json:"next-version"`
		NextVersionRound              int     `json:"next-version-round"`
		NextVersionSupported          bool    `json:"next-version-supported"`
		StoppedAtUnsupportedRound     bool    `json:"stopped-at-unsupported-round"`
		TimeSinceLastRound            int     `json:"time-since-last-round"`
		UpgradeDelay                  *int    `json:"upgrade-delay,omitempty"`
		UpgradeNextProtocolVoteBefore *int    `json:"upgrade-next-protocol-vote-before,omitempty"`
		UpgradeNoVotes                *int    `json:"upgrade-no-votes,omitempty"`
		UpgradeNodeVote               *bool   `json:"upgrade-node-vote,omitempty"`
		UpgradeVoteRounds             *int    `json:"upgrade-vote-rounds,omitempty"`
		UpgradeVotes                  *int    `json:"upgrade-votes,omitempty"`
		UpgradeVotesRequired          *int    `json:"upgrade-votes-required,omitempty"`
		UpgradeYesVotes               *int    `json:"upgrade-yes-votes,omitempty"`
	})
	data.LastRound = 10
	res := api.GetStatusResponse{
		Body:         nil,
		HTTPResponse: &httpResponse,
		JSON200:      data,
		JSON401:      nil,
		JSON500:      nil,
	}

	return &res, nil
}
