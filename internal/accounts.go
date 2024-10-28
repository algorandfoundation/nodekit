package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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

func getAddressesFromGenesis() ([]string, string, string, error) {

	// TODO: replace with calls to GetGenesis

	resp, err := http.Get("http://localhost:8080/genesis")
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

func isValidStatus(status string) bool {
	validStatuses := map[string]bool{
		"Online":            true,
		"Offline":           true,
		"Not Participating": true,
	}
	return validStatuses[status]
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
