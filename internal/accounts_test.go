package internal

import (
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

	// Create expectedAccounts dynamically from Online accounts, and mocked participation keys
	mockedPartKeys := make([]api.ParticipationKey, 0)
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

		mockedPartKeys = append(mockedPartKeys, api.ParticipationKey{
			Address:             address,
			EffectiveFirstValid: nil,
			EffectiveLastValid:  nil,
			Id:                  "",
			Key: api.AccountParticipation{
				SelectionParticipationKey: nil,
				StateProofKey:             nil,
				VoteParticipationKey:      nil,
				VoteFirstValid:            0,
				VoteLastValid:             9999999,
				VoteKeyDilution:           0,
			},
			LastBlockProposal: nil,
			LastStateProof:    nil,
			LastVote:          nil,
		})
	}

	// Mock StateModel
	state := &StateModel{
		ParticipationKeys: &mockedPartKeys,
	}

	// Call AccountsFromState
	accounts := AccountsFromState(state, client)

	// Assert results
	assert.Equal(t, expectedAccounts, accounts)

}
