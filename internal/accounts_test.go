package internal

import (
	"context"
	"testing"
	"time"

	"github.com/algorandfoundation/hack-tui/api"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
	"github.com/stretchr/testify/assert"
)

func Test_AccountsFromState(t *testing.T) {

	// Setup elevated client
	apiToken, err := securityprovider.NewSecurityProviderApiKey("header", "X-Algo-API-Token", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatal(err)
	}
	client, err := api.NewClientWithResponses("http://localhost:8080", api.WithRequestEditorFn(apiToken.Intercept))

	addresses, rewardsPool, feeSink, err := getAddressesFromGenesis(client)

	if err != nil {
		t.Fatal(err)
	}

	// Test getAccountOnlineStatus

	var mapAddressOnlineStatus = make(map[string]string)

	for _, address := range addresses {
		status, err := getAccountOnlineStatus(client, address)
		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, status == "Online" || status == "Offline")
		mapAddressOnlineStatus[address] = status
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

	// Test AccountFromState

	// Prepare expected results
	// Only include addresses with "Online" status
	onlineAddresses := make(map[string]string)
	for address, status := range mapAddressOnlineStatus {
		if status == "Online" {
			onlineAddresses[address] = status
		}
	}

	// Create expectedAccounts dynamically from Online accounts, corresponding to our part keys
	expectedAccounts := make(map[string]Account)
	for address, status := range onlineAddresses {
		expectedAccounts[address] = Account{
			Address:      address,
			Status:       status,
			Balance:      0,
			Expires:      time.Unix(0, 0),
			Keys:         1,
			LastModified: 0,
		}
	}

	// Get Part Keys
	// There should be at two online accounts in tuinet, so we can use them to test.
	partKeys, err := GetPartKeys(context.Background(), client)

	if err != nil {
		t.Fatal(err)
	}

	// Mock StateModel
	state := &StateModel{
		ParticipationKeys: partKeys,
	}

	// Call AccountsFromState
	accounts := AccountsFromState(state, client)

	// Assert results
	assert.Equal(t, expectedAccounts, accounts)

}
