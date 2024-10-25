package internal

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/algorandfoundation/hack-tui/api"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
	"github.com/stretchr/testify/assert"
)

func getAddressesFromGenesis(t *testing.T) ([]string, string, string) {

	// TODO: replace with calls to GetGenesis

	resp, err := http.Get("http://localhost:8080/genesis")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	// Check the response status code
	assert.Equal(t, 200, resp.StatusCode)

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	// Unmarshal the JSON response into a map
	var jsonResponse map[string]interface{}
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		t.Fatal(err)
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
					t.Logf("Address not found in entry: %+v", entry)
				}
			} else {
				t.Logf("Entry is not a map: %+v", entry)
			}
		}
	} else {
		t.Fatal("alloc is not a list")
	}

	if !rewardsPoolIncluded || !feeSinkIncluded {

		t.Fatalf("Expected RewardsPool and/or FeeSink addresses NOT found in genesis file")
	}

	return addresses, rewardsPool, feeSink
}

func isValidStatus(status string) bool {
	validStatuses := map[string]bool{
		"Online":            true,
		"Offline":           true,
		"Not Participating": true,
	}
	return validStatuses[status]
}

func Test_GetAccount(t *testing.T) {

	// Setup elevated client
	apiToken, err := securityprovider.NewSecurityProviderApiKey("header", "X-Algo-API-Token", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatal(err)
	}
	client, err := api.NewClientWithResponses("http://localhost:8080", api.WithRequestEditorFn(apiToken.Intercept))

	addresses, rewardsPool, feeSink := getAddressesFromGenesis(t)

	for _, address := range addresses {
		status, err := getAccountOnlineStatus(client, address)
		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, status == "Online" || status == "Offline")
	}

	status, err := getAccountOnlineStatus(client, rewardsPool)
	if err != nil {
		t.Fatal(err)
	}
	if status != "Not Participating" {
		t.Fatalf("Expected RewardsPool to be 'Not Participating', got %s", status)
	}

	status, err = getAccountOnlineStatus(client, feeSink)
	if err != nil {
		t.Fatal(err)
	}
	if status != "Not Participating" {
		t.Fatalf("Expected FeeSink to be 'Not Participating', got %s", status)
	}

	_, err = getAccountOnlineStatus(client, "invalid_address")
	if err == nil {
		t.Fatal("Expected error for invalid address")
	}

}
