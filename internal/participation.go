package internal

import (
	"context"
	"errors"
	"time"

	"github.com/algorandfoundation/hack-tui/api"
)

// GetPartKeys get the participation keys from the node
func GetPartKeys(ctx context.Context, client *api.ClientWithResponses) (*[]api.ParticipationKey, error) {
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
func ReadPartKey(ctx context.Context, client *api.ClientWithResponses, participationId string) (*api.ParticipationKey, error) {
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
	client *api.ClientWithResponses,
	keys *[]api.ParticipationKey,
	interval time.Duration,
) (*[]api.ParticipationKey, error) {
	// Fetch the latest keys
	currentKeys, err := GetPartKeys(ctx, client)
	if err != nil {
		return nil, err
	}
	// Check the length against known keys
	if len(*currentKeys) == len(*keys) {
		// Sleep then try again
		time.Sleep(interval)
		return waitForNewKey(ctx, client, keys, interval)
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
	return &participationKey, nil
}

// GenerateKeyPair creates a keypair and finds the result
func GenerateKeyPair(
	ctx context.Context,
	client *api.ClientWithResponses,
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
		return nil, errors.New(key.Status())
	}

	// Wait for the api to have a new key
	keys, err := waitForNewKey(ctx, client, originalKeys, 2*time.Second)
	if err != nil {
		return nil, err
	}

	// Find the new keypair in the results
	return findKeyPair(originalKeys, keys, address)
}

// DeletePartKey remove a key from the node
func DeletePartKey(ctx context.Context, client *api.ClientWithResponses, participationId string) error {
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
