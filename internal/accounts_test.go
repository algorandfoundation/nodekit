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

	params := api.GenerateParticipationKeysParams{
		Dilution: nil,
		First:    0,
		Last:     10000,
	}

	// Generate ParticipationKeys for all addresses
	var participationKeys []api.ParticipationKey
	for _, address := range addresses {
		key, err := GenerateKeyPair(context.Background(), client, address, &params)
		if err != nil {
			t.Fatal(err)
		}
		participationKeys = append(participationKeys, *key)
	}

	// Mock StateModel
	state := &StateModel{
		ParticipationKeys: &participationKeys,
	}

	// Call AccountsFromState
	accounts := AccountsFromState(state, client)

	// Create expectedAccounts dynamically
	expectedAccounts := make(map[string]Account)
	for _, address := range addresses {
		expectedAccounts[address] = Account{
			Address:      address,
			Status:       mapAddressOnlineStatus[address],
			Balance:      0,
			Expires:      time.Unix(0, 0),
			Keys:         1,
			LastModified: 0,
		}
	}

	// Assert results
	assert.Equal(t, expectedAccounts, accounts)

}
