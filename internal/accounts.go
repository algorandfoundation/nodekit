package internal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/algorandfoundation/hack-tui/api"
)

// Account represents a user's account, including address, status, balance, and number of keys.
type Account struct {
	Participation *api.AccountParticipation
	// Account Address is the algorand encoded address
	Address string
	// Status is the Online/Offline/"NotParticipating" status of the account
	Status string
	// Balance is the current holdings in ALGO for the address.
	// the balance should be tracked infrequently and use an appropriate distance from the
	// LastModified value.
	Balance int
	// A count of how many participation Keys exist on this node for this Account
	Keys int
	// Expires is the date the participation key will expire
	Expires time.Time
}

// Get Online Status of Account
func GetAccount(client api.ClientWithResponsesInterface, address string) (api.Account, error) {
	var format api.AccountInformationParamsFormat = "json"
	r, err := client.AccountInformationWithResponse(
		context.Background(),
		address,
		&api.AccountInformationParams{
			Format: &format,
		})

	var accountInfo api.Account
	if err != nil {
		return accountInfo, err
	}

	if r.StatusCode() != 200 {
		return accountInfo, errors.New(fmt.Sprintf("Failed to get account information. Received error code: %d", r.StatusCode()))
	}

	return *r.JSON200, nil
}

func GetExpiresTime(t Time, key api.ParticipationKey, state *StateModel) time.Time {
	now := t.Now()
	var expires = now.Add(-(time.Hour * 24 * 365 * 100))
	if key.LastBlockProposal != nil && state.Status.LastRound != 0 && state.Metrics.RoundTime != 0 {
		roundDiff := max(0, *key.EffectiveLastValid-int(state.Status.LastRound))
		distance := int(state.Metrics.RoundTime) * roundDiff
		expires = now.Add(time.Duration(distance))
	}
	return expires
}

// AccountsFromParticipationKeys maps an array of api.ParticipationKey to a keyed map of Account
func AccountsFromState(state *StateModel, t Time, client api.ClientWithResponsesInterface) map[string]Account {
	values := make(map[string]Account)
	if state == nil || state.ParticipationKeys == nil {
		return values
	}
	for _, key := range *state.ParticipationKeys {
		val, ok := values[key.Address]
		if !ok {
			var account = api.Account{
				Address: key.Address,
				Status:  "Unknown",
				Amount:  0,
			}
			if state.Status.State != SyncingState {
				var err error
				account, err = GetAccount(client, key.Address)
				// TODO: handle error
				if err != nil {
					// TODO: Logging
					panic(err)
				}
			}

			values[key.Address] = Account{
				Participation: account.Participation,
				Address:       key.Address,
				Status:        account.Status,
				Balance:       account.Amount / 1000000,
				Expires:       GetExpiresTime(t, key, state),
				Keys:          1,
			}
		} else {
			val.Keys++
			if val.Expires.Before(t.Now()) {
				now := t.Now()
				var expires = GetExpiresTime(t, key, state)
				if !expires.Before(now) {
					val.Expires = expires
				}
			}
			values[key.Address] = val
		}
	}

	return values
}

func ValidateAddress(address string) bool {
	_, err := types.DecodeAddress(address)
	if err != nil {
		return false
	}
	return true
}
