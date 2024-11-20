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
