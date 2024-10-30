package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
}

// Gets the list of addresses created at genesis from the genesis file
func getAddressesFromGenesis(client *api.ClientWithResponses) ([]string, string, string, error) {
	resp, err := client.GetGenesis(context.Background())
	if err != nil {
		return []string{}, "", "", err
	}

	if resp.StatusCode != 200 {
		return []string{}, "", "", errors.New(fmt.Sprintf("Failed to get genesis file. Received error code: %d", resp.StatusCode))
	}

	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []string{}, "", "", err
	}

	// Unmarshal the JSON response into a map
	var jsonResponse map[string]interface{}
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		return []string{}, "", "", err
	}

	// Two special addresses
	rewardsPool := "7777777777777777777777777777777777777777777777777774MSJUVU"
	feeSink := "A7NMWS3NT3IUDMLVO26ULGXGIIOUQ3ND2TXSER6EBGRZNOBOUIQXHIBGDE"
	rewardsPoolIncluded := false
	feeSinkIncluded := false

	// Loop over each entry in the "alloc" list and collect the "addr" values
	var addresses []string
	if allocList, ok := jsonResponse["alloc"].([]interface{}); ok {
		for _, entry := range allocList {
			if entryMap, ok := entry.(map[string]interface{}); ok {
				if addr, ok := entryMap["addr"].(string); ok {
					if addr == rewardsPool {
						rewardsPoolIncluded = true
					} else if addr == feeSink {
						feeSinkIncluded = true
					} else {
						addresses = append(addresses, addr)
					}
				} else {
					return []string{}, "", "", fmt.Errorf("In genesis.json no addr string found in list element entry:  %+v", entry)
				}
			} else {
				return []string{}, "", "", fmt.Errorf("In genesis.json list element of alloc-field is not a map:  %+v", entry)
			}
		}
	} else {
		return []string{}, "", "", errors.New("alloc is not a list")
	}

	if !rewardsPoolIncluded || !feeSinkIncluded {
		return []string{}, "", "", errors.New("Expected RewardsPool and/or FeeSink addresses NOT found in genesis file")
	}

	return addresses, rewardsPool, feeSink, nil
}

// Get Online Status of Account
func GetAccount(client *api.ClientWithResponses, address string) (api.Account, error) {
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

// AccountsFromParticipationKeys maps an array of api.ParticipationKey to a keyed map of Account
func AccountsFromState(state *StateModel, t Time, client *api.ClientWithResponses) map[string]Account {
	values := make(map[string]Account)
	if state == nil || state.ParticipationKeys == nil {
		return values
	}
	for _, key := range *state.ParticipationKeys {
		val, ok := values[key.Address]
		if !ok {

			account, err := GetAccount(client, key.Address)

			// TODO: handle error
			if err != nil {
				// TODO: Logging
				panic(err)
			}

			var expires = t.Now()
			if key.EffectiveLastValid != nil {
				now := t.Now()
				roundDiff := max(0, *key.EffectiveLastValid-int(state.Status.LastRound))
				distance := int(state.Metrics.RoundTime) * roundDiff
				expires = now.Add(time.Duration(distance))
			}

			values[key.Address] = Account{
				Address: key.Address,
				Status:  account.Status,
				Balance: account.Amount / 1000000,
				Expires: expires,
				Keys:    1,
			}
		} else {
			val.Keys++
			values[key.Address] = val
		}
	}

	return values
}
