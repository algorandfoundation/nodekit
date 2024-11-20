package test

import (
	"context"
	"errors"
	"github.com/algorandfoundation/hack-tui/api"
	"github.com/algorandfoundation/hack-tui/internal/test/mock"
	"net/http"
)

func GetClient(throws bool) api.ClientWithResponsesInterface {
	return NewClient(throws)
}

type Client struct {
	api.ClientWithResponsesInterface
	Errors bool
}

func NewClient(throws bool) api.ClientWithResponsesInterface {
	client := new(Client)
	if throws {
		client.Errors = true
	}
	return client
}
func (c *Client) GetParticipationKeysWithResponse(ctx context.Context, reqEditors ...api.RequestEditorFn) (*api.GetParticipationKeysResponse, error) {
	httpResponse := http.Response{StatusCode: 200}
	clone := mock.Keys
	res := api.GetParticipationKeysResponse{
		Body:         nil,
		HTTPResponse: &httpResponse,
		JSON200:      &clone,
		JSON400:      nil,
		JSON401:      nil,
		JSON404:      nil,
		JSON500:      nil,
	}
	if c.Errors {
		return nil, errors.New("test error")
	}
	return &res, nil
}

func (c *Client) DeleteParticipationKeyByIDWithResponse(ctx context.Context, participationId string, reqEditors ...api.RequestEditorFn) (*api.DeleteParticipationKeyByIDResponse, error) {
	httpResponse := http.Response{StatusCode: 200}
	res := api.DeleteParticipationKeyByIDResponse{
		Body:         nil,
		HTTPResponse: &httpResponse,
		JSON400:      nil,
		JSON401:      nil,
		JSON404:      nil,
		JSON500:      nil,
	}

	if c.Errors {
		return nil, errors.New("test error")
	}
	return &res, nil
}

func (c *Client) GenerateParticipationKeysWithResponse(ctx context.Context, address string, params *api.GenerateParticipationKeysParams, reqEditors ...api.RequestEditorFn) (*api.GenerateParticipationKeysResponse, error) {
	mock.Keys = append(mock.Keys, api.ParticipationKey{
		Address:             "",
		EffectiveFirstValid: nil,
		EffectiveLastValid:  nil,
		Id:                  "",
		Key:                 api.AccountParticipation{},
		LastBlockProposal:   nil,
		LastStateProof:      nil,
		LastVote:            nil,
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
	httpResponse := http.Response{StatusCode: 200}
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
	res := api.GetVersionResponse{
		Body:         nil,
		HTTPResponse: &httpResponse,
		JSON200:      &version,
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
