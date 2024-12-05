package test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/algorandfoundation/algorun-tui/api"
	"io"
)

// GetAddressesFromGenesis gets the list of addresses created at genesis from the genesis file
func GetAddressesFromGenesis(ctx context.Context, client api.ClientInterface) ([]string, string, string, error) {
	resp, err := client.GetGenesis(ctx)
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
