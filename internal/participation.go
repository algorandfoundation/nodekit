package internal

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/algorandfoundation/hack-tui/api"
)

// GetPartKeys get the participation keys from the node
func GetPartKeys(ctx context.Context, client api.ClientWithResponsesInterface) (*[]api.ParticipationKey, error) {
	parts, err := client.GetParticipationKeysWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	if parts.StatusCode() != 200 {
		return nil, errors.New(parts.Status())
	}
	return parts.JSON200, err
}

// ReadPartKey get a specific participation key by id
func ReadPartKey(ctx context.Context, client api.ClientWithResponsesInterface, participationId string) (*api.ParticipationKey, error) {
	key, err := client.GetParticipationKeyByIDWithResponse(ctx, participationId)
	if err != nil {
		return nil, err
	}
	if key.StatusCode() != 200 {
		return nil, errors.New(key.Status())
	}
	return key.JSON200, err
}

// waitForNewKey await the new key based on known existing keys
// We should try to update the API endpoint
func waitForNewKey(
	ctx context.Context,
	client api.ClientWithResponsesInterface,
	keys *[]api.ParticipationKey,
	interval time.Duration,
	timeout time.Duration,
) (*[]api.ParticipationKey, error) {
	if timeout <= 0*time.Second {
		return nil, errors.New("timeout occurred waiting for new key")
	}
	timeout = timeout - interval
	// Fetch the latest keys
	currentKeys, err := GetPartKeys(ctx, client)
	if err != nil {
		return nil, err
	}
	if keys == nil && currentKeys != nil {
		return currentKeys, nil
	}
	// Check the length against known keys
	if currentKeys == nil || len(*currentKeys) == 0 || len(*currentKeys) == len(*keys) {
		// Sleep then try again
		time.Sleep(interval)
		return waitForNewKey(ctx, client, keys, interval, timeout)
	}
	return currentKeys, nil
}

// findKeyPair look for a new key based on address between two key lists
// this is not robust, and we should try to update the API endpoint to wait for
// the key creation and return its metadata to the caller
func findKeyPair(
	originalKeys *[]api.ParticipationKey,
	currentKeys *[]api.ParticipationKey,
	address string,
) (*api.ParticipationKey, error) {
	var participationKey api.ParticipationKey
	for _, key := range *currentKeys {
		if key.Address == address {
			for _, oKey := range *originalKeys {
				if oKey.Id != key.Id {
					participationKey = key
				}
			}
		}
	}
	// If keys are empty, return the found keys
	if originalKeys == nil || len(*originalKeys) == 0 {
		keys := *currentKeys
		participationKey = keys[0]
	}
	if participationKey.Id == "" {
		return nil, errors.New("key not found")
	}
	return &participationKey, nil
}

// GenerateKeyPair creates a keypair and finds the result
func GenerateKeyPair(
	ctx context.Context,
	client api.ClientWithResponsesInterface,
	address string,
	params *api.GenerateParticipationKeysParams,
) (*api.ParticipationKey, error) {
	// The api response is an empty body, we need to fetch known keys first
	originalKeys, err := GetPartKeys(ctx, client)
	if err != nil {
		return nil, err
	}
	// Generate a new keypair
	key, err := client.GenerateParticipationKeysWithResponse(ctx, address, params)
	if err != nil {
		return nil, err
	}
	if key.StatusCode() != 200 {
		return nil, errors.New("something went wrong")
	}

	// Wait for the api to have a new key
	keys, err := waitForNewKey(ctx, client, originalKeys, 2*time.Second, 20*time.Minute)
	if err != nil {
		return nil, err
	}

	// Find the new keypair in the results
	return findKeyPair(originalKeys, keys, address)
}

// DeletePartKey remove a key from the node
func DeletePartKey(ctx context.Context, client api.ClientWithResponsesInterface, participationId string) error {
	deletion, err := client.DeleteParticipationKeyByIDWithResponse(ctx, participationId)
	if err != nil {
		return err
	}
	if deletion.StatusCode() != 200 {
		return errors.New(deletion.Status())
	}
	return nil
}

// Removes a participation key from the list of keys
func RemovePartKeyByID(slice *[]api.ParticipationKey, id string) {
	for i, item := range *slice {
		if item.Id == id {
			*slice = append((*slice)[:i], (*slice)[i+1:]...)
			return
		}
	}
}

func FindParticipationIdForVoteKey(slice *[]api.ParticipationKey, votekey []byte) *string {
	for _, item := range *slice {
		if string(item.Key.VoteParticipationKey) == string(votekey) {
			return &item.Id
		}
	}
	return nil
}

func ToLoraDeepLink(network string, offline bool, part api.ParticipationKey) (string, error) {
	fee := 2000000
	var loraNetwork = strings.Replace(strings.Replace(network, "-v1.0", "", 1), "-v1", "", 1)
	if loraNetwork == "dockernet" || loraNetwork == "tuinet" {
		loraNetwork = "localnet"
	}

	var query = ""
	idx := url.QueryEscape("[0]")
	if offline {
		query = fmt.Sprintf(
			"type[0]=keyreg&sender[0]=%s",
			part.Address,
		)
	} else {
		query = fmt.Sprintf(
			"type[0]=keyreg&fee[0]=%d&sender[0]=%s&selkey[0]=%s&sprfkey[0]=%s&votekey[0]=%s&votefst[0]=%d&votelst[0]=%d&votekd[0]=%d",
			fee,
			part.Address,
			base64.RawURLEncoding.EncodeToString(part.Key.SelectionParticipationKey),
			base64.RawURLEncoding.EncodeToString(*part.Key.StateProofKey),
			base64.RawURLEncoding.EncodeToString(part.Key.VoteParticipationKey),
			part.Key.VoteFirstValid,
			part.Key.VoteLastValid,
			part.Key.VoteKeyDilution,
		)
	}
	return fmt.Sprintf("https://lora.algokit.io/%s/transaction-wizard?%s", loraNetwork, strings.Replace(query, "[0]", idx, -1)), nil
}
