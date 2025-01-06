package participation

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/algorandfoundation/algorun-tui/api"
)

// RangeType represents a type of range, such as time-based or round-based, used in participation key generation.
type RangeType string

const (

	// TimeRange defines a range type measured in seconds, used for specifying time-based participation key validity.
	TimeRange RangeType = "seconds"

	// RoundRange represents a range type defined in terms of rounds for participation key validity.
	RoundRange RangeType = "rounds"
)

// List represents a collection of api.ParticipationKey elements.
type List []api.ParticipationKey

// GetList get the participation keys from the node
func GetList(ctx context.Context, client api.ClientWithResponsesInterface) (List, api.ResponseInterface, error) {
	partKeysResponse, err := client.GetParticipationKeysWithResponse(ctx)
	if err != nil {
		return nil, partKeysResponse, err
	}
	if partKeysResponse.StatusCode() != 200 {
		return nil, partKeysResponse, errors.New(partKeysResponse.Status())
	}
	return *partKeysResponse.JSON200, partKeysResponse, err
}

// GetKey get a specific participation key by id
func GetKey(ctx context.Context, client api.ClientWithResponsesInterface, participationId string) (*api.ParticipationKey, api.ResponseInterface, error) {
	keyResponse, err := client.GetParticipationKeyByIDWithResponse(ctx, participationId)
	if err != nil {
		return nil, keyResponse, err
	}
	if keyResponse.StatusCode() != 200 {
		return nil, keyResponse, errors.New(keyResponse.Status())
	}
	return keyResponse.JSON200, keyResponse, err
}

// GenerateKeys creates a participation keypair for the specified address and validity period parameters.
// It ensures the key creation on the node is complete and verifies the key's properties before returning it.
// If the key creation fails, or the operation is interrupted/timed out, it returns an appropriate error.
func GenerateKeys(
	ctx context.Context,
	client api.ClientWithResponsesInterface,
	address string,
	params *api.GenerateParticipationKeysParams,
) (*api.ParticipationKey, error) {
	// Generate a new keypair
	key, err := client.GenerateParticipationKeysWithResponse(ctx, address, params)
	if err != nil {
		return nil, err
	}
	if key.StatusCode() != 200 {
		if key.JSON400 != nil {
			return nil, errors.New(key.JSON400.Message)
		}

		status := key.Status()
		if status != "" {
			return nil, errors.New(status)
		}
		return nil, errors.New("something went wrong")
	}

	// 😠 - Zero 2024
	for {
		select {
		case <-ctx.Done():
			return nil, context.Canceled
		case <-time.After(2 * time.Second):
			partKeys, _, err := GetList(ctx, client)
			if partKeys == nil || err != nil {
				return nil, errors.New("failed to get participation keys")
			}
			for _, k := range partKeys {
				if k.Address == address &&
					k.Key.VoteFirstValid == params.First &&
					k.Key.VoteLastValid == params.Last {
					return &k, nil
				}
			}
		case <-time.After(20 * time.Minute):
			return nil, errors.New("timeout waiting for key to be created")
		}
	}
}

// Delete remove a key from the node
func Delete(ctx context.Context, client api.ClientWithResponsesInterface, participationId string) error {
	deletion, err := client.DeleteParticipationKeyByIDWithResponse(ctx, participationId)
	if err != nil {
		return err
	}
	if deletion.StatusCode() != 200 {
		return errors.New(deletion.Status())
	}
	return nil
}

// RemovePartKeyByID Removes a participation key from the list of keys
func RemovePartKeyByID(slice *List, id string) {
	for i, item := range *slice {
		if item.Id == id {
			*slice = append((*slice)[:i], (*slice)[i+1:]...)
			return
		}
	}
}

// FindParticipationIdForVoteKey searches a List for an item's VoteParticipationKey matching the given votekey and returns its ID.
// It returns a pointer to the ID as a string if a match is found, or nil otherwise.
func FindParticipationIdForVoteKey(slice List, votekey []byte) *string {
	for _, item := range slice {
		if string(item.Key.VoteParticipationKey) == string(votekey) {
			return &item.Id
		}
	}
	return nil
}

// IsActive checks if the given participation key matches the account's registered participation details and is valid.
func IsActive(part api.ParticipationKey, account api.AccountParticipation) bool {
	var equal = false
	if bytes.Equal(part.Key.VoteParticipationKey, account.VoteParticipationKey) &&
		part.Key.VoteLastValid == account.VoteLastValid &&
		part.Key.VoteFirstValid == account.VoteFirstValid {
		equal = true
	}
	return equal
}

// OnlineShortLinkBody represents the request payload for creating an online short link.
// It contains account details, cryptographic keys in Base64, validity range, key dilution, and network information.
type OnlineShortLinkBody struct {
	Account          string `json:"account"`
	VoteKeyB64       string `json:"voteKeyB64"`
	SelectionKeyB64  string `json:"selectionKeyB64"`
	StateProofKeyB64 string `json:"stateProofKeyB64"`
	VoteFirstValid   int    `json:"voteFirstValid"`
	VoteLastValid    int    `json:"voteLastValid"`
	KeyDilution      int    `json:"keyDilution"`
	Network          string `json:"network"`
}

// GetOnlineShortLink sends a POST request to create an online short link
// and returns the response or an error if it occurs.
func GetOnlineShortLink(http api.HttpPkgInterface, part OnlineShortLinkBody) (ShortLinkResponse, error) {
	var response ShortLinkResponse
	data, err := json.Marshal(part)
	if err != nil {
		return response, err
	}
	res, err := http.Post("http://b.nodekit.run/online", "applicaiton/json", bytes.NewReader(data))
	if err != nil {
		return response, err
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return response, err
	}

	return response, nil
}

// ShortLinkResponse represents the response structure for a shortened link,
// containing its unique identifier.
type ShortLinkResponse struct {
	Id string `json:"id"`
}

// OfflineShortLinkBody represents the request body for creating an
// offline short link containing an address and network.
type OfflineShortLinkBody struct {
	Account string `json:"account"`
	Network string `json:"network"`
}

// GetOfflineShortLink sends an OnlineShortLinkBody to create an offline short link and returns the corresponding response.
// Uses the provided HttpPkgInterface for the POST request and handles JSON encoding/decoding of request and response.
// Returns an OfflineShortLinkResponse on success or an error if the operation fails.
func GetOfflineShortLink(http api.HttpPkgInterface, offline OfflineShortLinkBody) (ShortLinkResponse, error) {
	var response ShortLinkResponse
	data, err := json.Marshal(offline)
	if err != nil {
		return response, err
	}
	res, err := http.Post("http://b.nodekit.run/offline", "applicaiton/json", bytes.NewReader(data))
	if err != nil {
		return response, err
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&response)
	if err != nil {
		return response, err
	}

	return response, nil
}

// ToShortLink generates a shortened URL string using the unique
// identifier from the provided ShortLinkResponse.
func ToShortLink(link ShortLinkResponse) string {
	return fmt.Sprintf("https://b.nodekit.run/%s", link.Id)
}
