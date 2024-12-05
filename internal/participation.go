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

// GenerateKeyPair creates a keypair and finds the result
func GenerateKeyPair(
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
		status := key.Status()
		if status != "" {
			return nil, errors.New(status)
		}
		return nil, errors.New("something went wrong")
	}
	for {
		select {
		case <-ctx.Done():
			return nil, context.Canceled
		case <-time.After(2 * time.Second):
			partKeys, err := GetPartKeys(ctx, client)
			if partKeys == nil || err != nil {
				return nil, errors.New("failed to get participation keys")
			}
			for _, k := range *partKeys {
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

func ToLoraDeepLink(network string, offline bool, incentiveEligible bool, part api.ParticipationKey) (string, error) {
	var loraNetwork = strings.Replace(strings.Replace(network, "-v1.0", "", 1), "-v1", "", 1)
	if loraNetwork == "dockernet" || loraNetwork == "tuinet" {
		loraNetwork = "localnet"
	}

	var query = ""
	encodedIndex := url.QueryEscape("[0]")
	if offline {
		query = fmt.Sprintf(
			"type[0]=keyreg&sender[0]=%s",
			part.Address,
		)
	} else {
		query = fmt.Sprintf(
			"type[0]=keyreg&sender[0]=%s&selkey[0]=%s&sprfkey[0]=%s&votekey[0]=%s&votefst[0]=%d&votelst[0]=%d&votekd[0]=%d",
			part.Address,
			base64.RawURLEncoding.EncodeToString(part.Key.SelectionParticipationKey),
			base64.RawURLEncoding.EncodeToString(*part.Key.StateProofKey),
			base64.RawURLEncoding.EncodeToString(part.Key.VoteParticipationKey),
			part.Key.VoteFirstValid,
			part.Key.VoteLastValid,
			part.Key.VoteKeyDilution,
		)
		if incentiveEligible {
			// TODO: enable fee with either feature flag or config flag
			// query += fmt.Sprintf("&fee[0]=%d", 2000000)
		}
	}
	return fmt.Sprintf("https://lora.algokit.io/%s/transaction-wizard?%s", loraNetwork, strings.Replace(query, "[0]", encodedIndex, -1)), nil
}
