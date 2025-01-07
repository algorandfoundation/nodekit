package algod

import (
	"context"
	"errors"
	"fmt"
	"github.com/algorandfoundation/nodekit/internal/algod/participation"
	"github.com/algorandfoundation/nodekit/internal/algod/utils"
	"github.com/algorandfoundation/nodekit/internal/system"
	"time"

	"github.com/algorand/go-algorand-sdk/v2/types"
	"github.com/algorandfoundation/nodekit/api"
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

// ParticipationKeysToAccounts converts a slice of ParticipationKey objects into a map of Account objects.
// The keys parameter is a slice of pointers to ParticipationKey instances.
// The prev parameter is an optional map that allows merging of existing accounts with new ones.
// Returns a map where each key is an address from a ParticipationKey, and the value is a corresponding Account.
func ParticipationKeysToAccounts(keys []api.ParticipationKey) map[string]Account {
	// Allow merging of existing accounts
	var accounts = make(map[string]Account)

	// Must have keys to process
	if keys == nil {
		return accounts
	}

	// Add missing Accounts
	for _, key := range keys {
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

// Merge updates the Account instance with data from the provided api.Account and returns the updated Account.
// It updates fields such as Status, Balance, Participation, and IncentiveEligible based on the rpcAccount values.
func (a Account) Merge(rpcAccount api.Account) Account {
	a.Status = rpcAccount.Status
	a.Balance = rpcAccount.Amount / 1000000
	a.Participation = rpcAccount.Participation

	var incentiveEligible = false
	if rpcAccount.IncentiveEligible == nil {
		incentiveEligible = false
	} else {
		incentiveEligible = *rpcAccount.IncentiveEligible
	}

	a.IncentiveEligible = incentiveEligible

	if rpcAccount.Participation != nil {
		a.Participation = rpcAccount.Participation
	}

	return a
}

// GetExpiresTime calculates the expiration time of the account's participation key based on round differences and duration.
// Returns nil if the account has no participation or if the expiration time cannot be determined.
func (a Account) GetExpiresTime(t system.Time, lastRound int, roundTime time.Duration) *time.Time {
	if a.Participation == nil {
		return nil
	}
	return utils.GetExpiresTime(t, lastRound, roundTime, a.Participation.VoteLastValid)
}

// UpdateExpiredTime updates the account's expiration time and identifies if the account has a non-resident participation key.
// It checks if the account is offline or if its local participation key matches one of the provided keys.
// The method recalculates the expiration time based on the last round and round duration.
func (a Account) UpdateExpiredTime(t system.Time, keys []api.ParticipationKey, lastRound int, roundTime time.Duration) Account {
	var nonResidentKey = true
	for _, key := range keys {
		// We have the key locally, update the residency
		if a.Status == "Offline" || (key.Address == a.Address && a.Participation != nil && participation.IsActive(key, *a.Participation)) {
			nonResidentKey = false
		}
	}
	a.NonResidentKey = nonResidentKey
	a.Expires = a.GetExpiresTime(t, lastRound, roundTime)
	return a
}

// ValidateAddress checks the validity of an Algorand address by decoding it. Returns true for valid addresses, false otherwise.
func ValidateAddress(address string) bool {
	_, err := types.DecodeAddress(address)
	if err != nil {
		return false
	}
	return true
}
