package internal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/algorandfoundation/hack-tui/api"
)

// Account represents a user's account, including address, status, balance, and number of keys.
type Account struct {
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
	// The LastModified round, this only pertains to keys that can be updated
	LastModified int
}

// Get Online Status of Account
func getAccountOnlineStatus(client *api.ClientWithResponses, address string) (string, error) {
	var format api.AccountInformationParamsFormat = "json"
	r, err := client.AccountInformationWithResponse(
		context.Background(),
		address,
		&api.AccountInformationParams{
			Format: &format,
		})

	if err != nil {
		return "N/A", err
	}

	if r.StatusCode() != 200 {
		return "N/A", errors.New(fmt.Sprintf("Failed to get account information. Received error code: %d", r.StatusCode()))
	}

	if r.JSON200 == nil {
		return "N/A", errors.New("Failed to get account information. JSON200 is nil")
	}

	return r.JSON200.Status, nil
}

// AccountsFromParticipationKeys maps an array of api.ParticipationKey to a keyed map of Account
func AccountsFromState(state *StateModel, client *api.ClientWithResponses) map[string]Account {
	values := make(map[string]Account)
	if state == nil || state.ParticipationKeys == nil {
		return values
	}
	for _, key := range *state.ParticipationKeys {
		val, ok := values[key.Address]
		if !ok {

			statusOnline, err := getAccountOnlineStatus(client, key.Address)

			if err != nil {
				// TODO: Logging
				panic(err)
			}

			values[key.Address] = Account{
				Address: key.Address,
				Status:  statusOnline,
				Balance: 0,
				Expires: time.Unix(0, 0),
				Keys:    1,
			}
		} else {
			val.Keys++
			values[key.Address] = val
		}
	}

	return values
}
