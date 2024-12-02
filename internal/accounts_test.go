package internal

import (
	"context"
	"github.com/algorandfoundation/hack-tui/api"
	"github.com/algorandfoundation/hack-tui/internal/test"
	"github.com/algorandfoundation/hack-tui/internal/test/mock"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_AccountsFromState(t *testing.T) {

	// Setup elevated client
	apiToken, err := securityprovider.NewSecurityProviderApiKey("header", "X-Algo-API-Token", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatal(err)
	}
	client, err := api.NewClientWithResponses("http://localhost:8080", api.WithRequestEditorFn(apiToken.Intercept))

	addresses, rewardsPool, feeSink, err := test.GetAddressesFromGenesis(context.Background(), client)

	if err != nil {
		t.Fatal(err)
	}

	var mapAccounts = make(map[string]api.Account)
	var onlineAccounts = make([]api.Account, 0)
	for _, address := range addresses {
		acct, err := GetAccount(client, address)
		if err != nil {
			t.Fatal(err)
		}

		assert.True(t, acct.Status == "Online" || acct.Status == "Offline")
		mapAccounts[address] = acct
		if acct.Status == "Online" {
			onlineAccounts = append(onlineAccounts, acct)
		}
	}

	acct, err := GetAccount(client, rewardsPool)
	if err != nil {
		t.Fatal(err)
	}
	if acct.Status != "Not Participating" {
		t.Fatalf("Expected RewardsPool to be 'Not Participating', got %s", acct.Status)
	}

	acct, err = GetAccount(client, feeSink)
	if err != nil {
		t.Fatal(err)
	}
	if acct.Status != "Not Participating" {
		t.Fatalf("Expected FeeSink to be 'Not Participating', got %s", acct.Status)
	}

	_, err = GetAccount(client, "invalid_address")
	if err == nil {
		t.Fatal("Expected error for invalid address")
	}

	// Test Account from State

	effectiveFirstValid := 0
	effectiveLastValid := 10000
	lastProposedRound := 1336
	// Create mockedPart Keys
	var mockedPartKeys = []api.ParticipationKey{
		{
			Address:             onlineAccounts[0].Address,
			EffectiveFirstValid: &effectiveFirstValid,
			EffectiveLastValid:  &effectiveLastValid,
			Id:                  "",
			Key: api.AccountParticipation{
				SelectionParticipationKey: nil,
				StateProofKey:             nil,
				VoteParticipationKey:      nil,
				VoteFirstValid:            0,
				VoteLastValid:             9999999,
				VoteKeyDilution:           0,
			},
			LastBlockProposal: &lastProposedRound,
			LastStateProof:    nil,
			LastVote:          nil,
		},
		{
			Address:             onlineAccounts[0].Address,
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
		},
		{
			Address:             onlineAccounts[1].Address,
			EffectiveFirstValid: &effectiveFirstValid,
			EffectiveLastValid:  &effectiveLastValid,
			Id:                  "",
			Key: api.AccountParticipation{
				SelectionParticipationKey: nil,
				StateProofKey:             nil,
				VoteParticipationKey:      nil,
				VoteFirstValid:            0,
				VoteLastValid:             9999999,
				VoteKeyDilution:           0,
			},
			LastBlockProposal: &lastProposedRound,
			LastStateProof:    nil,
			LastVote:          nil,
		},
	}

	// Mock StateModel
	state := &StateModel{
		Metrics: MetricsModel{
			Enabled:   true,
			Window:    100,
			RoundTime: time.Duration(2) * time.Second,
			TPS:       20,
			RX:        1024,
			TX:        2048,
		},
		Status: StatusModel{
			State:       "WATCHING",
			Version:     "v0.0.0-test",
			Network:     "tuinet",
			Voting:      false,
			NeedsUpdate: false,
			LastRound:   1337,
		},
		ParticipationKeys: &mockedPartKeys,
	}

	// Calculate expiration
	clock := new(mock.Clock)
	now := clock.Now()
	roundDiff := max(0, effectiveLastValid-int(state.Status.LastRound))
	distance := int(state.Metrics.RoundTime) * roundDiff
	expires := now.Add(time.Duration(distance))

	// Construct expected accounts
	expectedAccounts := map[string]Account{
		onlineAccounts[0].Address: {
			Participation: onlineAccounts[0].Participation,
			Address:       onlineAccounts[0].Address,
			Status:        onlineAccounts[0].Status,
			Balance:       onlineAccounts[0].Amount / 1_000_000,
			Keys:          2,
			Expires:       expires,
		},
		onlineAccounts[1].Address: {
			Participation: onlineAccounts[1].Participation,
			Address:       onlineAccounts[1].Address,
			Status:        onlineAccounts[1].Status,
			Balance:       onlineAccounts[1].Amount / 1_000_000,
			Keys:          1,
			Expires:       expires,
		},
	}

	// Call AccountsFromState
	accounts := AccountsFromState(state, clock, client)

	// Assert results
	assert.Equal(t, expectedAccounts, accounts)

}
