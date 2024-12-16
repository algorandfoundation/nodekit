package internal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"time"

	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/algorandfoundation/algorun-tui/api"
)

// Account represents a user's account, including address, status, balance, and number of keys.
type Account struct {
	Participation *api.AccountParticipation
	// IncentiveEligible determines the minimum fee
	IncentiveEligible bool
	// NonResidentKey finds an online account that is missing locally
	NonResidentKey bool
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
	Expires *time.Time
}

// GetAccount status of api.Account
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

// GetExpiresTime calculates and returns the expiration time for a participation key based on the current account state.
func GetExpiresTime(t Time, lastRound int, roundTime time.Duration, account Account) *time.Time {
	now := t.Now()
	var expires time.Time
	if account.Status == "Online" &&
		account.Participation != nil &&
		lastRound != 0 &&
		roundTime != 0 {
		roundDiff := max(0, account.Participation.VoteLastValid-int(lastRound))
		distance := int(roundTime) * roundDiff
		expires = now.Add(time.Duration(distance))
		return &expires
	}
	return nil
}

// ParticipationKeysToAccounts converts a slice of ParticipationKey objects into a map of Account objects.
// The keys parameter is a slice of pointers to ParticipationKey instances.
// The prev parameter is an optional map that allows merging of existing accounts with new ones.
// Returns a map where each key is an address from a ParticipationKey, and the value is a corresponding Account.
func ParticipationKeysToAccounts(keys *[]api.ParticipationKey) map[string]Account {
	// Allow merging of existing accounts
	var accounts = make(map[string]Account)

	// Must have keys to process
	if keys == nil {
		return accounts
	}

	// Add missing Accounts
	for _, key := range *keys {
		if _, ok := accounts[key.Address]; !ok {
			accounts[key.Address] = Account{
				Participation:     nil,
				IncentiveEligible: false,
				Address:           key.Address,
				Status:            "Unknown",
				Balance:           0,
				Keys:              1,
				Expires:           nil,
			}
		} else {
			acct := accounts[key.Address]
			acct.Keys++
			accounts[key.Address] = acct
		}
	}
	return accounts
}

func UpdateAccountFromRPC(account Account, rpcAccount api.Account) Account {
	account.Status = rpcAccount.Status
	account.Balance = rpcAccount.Amount / 1000000
	account.Participation = rpcAccount.Participation

	var incentiveEligible = false
	if rpcAccount.IncentiveEligible == nil {
		incentiveEligible = false
	} else {
		incentiveEligible = *rpcAccount.IncentiveEligible
	}

	account.IncentiveEligible = incentiveEligible

	return account
}

func IsParticipationKeyActive(part api.ParticipationKey, account api.AccountParticipation) bool {
	var equal = false
	if bytes.Equal(part.Key.VoteParticipationKey, account.VoteParticipationKey) &&
		part.Key.VoteLastValid == account.VoteLastValid &&
		part.Key.VoteFirstValid == account.VoteFirstValid {
		equal = true
	}
	return equal
}

func UpdateAccountExpiredTime(t Time, account Account, state *StateModel) Account {
	var nonResidentKey = true
	for _, key := range *state.ParticipationKeys {
		// We have the key locally, update the residency
		if account.Status == "Offline" || (key.Address == account.Address && account.Participation != nil && IsParticipationKeyActive(key, *account.Participation)) {
			nonResidentKey = false
		}
	}
	account.NonResidentKey = nonResidentKey
	account.Expires = GetExpiresTime(t, int(state.Status.LastRound), state.Metrics.RoundTime, account)
	return account
}

// AccountsFromState maps an array of api.ParticipationKey to a keyed map of Account
func AccountsFromState(state *StateModel, t Time, client api.ClientWithResponsesInterface) (map[string]Account, error) {
	if state == nil {
		return make(map[string]Account), nil
	}

	accounts := ParticipationKeysToAccounts(state.ParticipationKeys)

	for _, acct := range accounts {
		// For each account, update the data from the RPC endpoint
		if state.Status.State != algod.SyncingState {
			rpcAcct, err := GetAccount(client, acct.Address)
			if err != nil {
				return nil, err
			}
			accounts[acct.Address] = UpdateAccountFromRPC(acct, rpcAcct)
			accounts[acct.Address] = UpdateAccountExpiredTime(t, accounts[acct.Address], state)
		}
	}

	return accounts, nil
}

func ValidateAddress(address string) bool {
	_, err := types.DecodeAddress(address)
	if err != nil {
		return false
	}
	return true
}
