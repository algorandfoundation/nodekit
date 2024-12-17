package participation

import (
	"context"
	"fmt"
	"github.com/algorandfoundation/algorun-tui/api"
	"github.com/algorandfoundation/algorun-tui/internal/test"
	"github.com/algorandfoundation/algorun-tui/internal/test/mock"
	"testing"
)

func Test_ToLoraDeeplink(t *testing.T) {
	link, err := ToLoraDeepLink("tuinet-v1", true, true, api.ParticipationKey{
		Address:             "ABC",
		EffectiveFirstValid: nil,
		EffectiveLastValid:  nil,
		Id:                  "",
		Key:                 api.AccountParticipation{},
		LastBlockProposal:   nil,
		LastStateProof:      nil,
		LastVote:            nil,
	})
	if err != nil {
		t.Error(err)
	}
	if link != "https://lora.algokit.io/localnet/transaction-wizard?type%5B0%5D=keyreg&sender%5B0%5D=ABC" {
		t.Error("Link should be a known deeplink")
	}

	// TODO put back
	// link, err = ToLoraDeepLink("tuinet-v1", false, true, mock.Keys[0])
	// if err != nil {
	// 	t.Error(err)
	// }
	// if link != "https://lora.algokit.io/localnet/transaction-wizard?type%5B0%5D=keyreg&sender%5B0%5D=ABC&selkey%5B0%5D=VEVTVEtFWQ&sprfkey%5B0%5D=VEVTVEtFWQ&votekey%5B0%5D=VEVTVEtFWQ&votefst%5B0%5D=0&votelst%5B0%5D=30000&votekd%5B0%5D=100&fee%5B0%5D=2000000" {
	// 	t.Error("Link should be a known deeplink fee")
	// }

	link, err = ToLoraDeepLink("tuinet-v1", false, false, mock.Keys[0])
	if err != nil {
		t.Error(err)
	}
	if link != "https://lora.algokit.io/localnet/transaction-wizard?type%5B0%5D=keyreg&sender%5B0%5D=ABC&selkey%5B0%5D=VEVTVEtFWQ&sprfkey%5B0%5D=VEVTVEtFWQ&votekey%5B0%5D=VEVTVEtFWQ&votefst%5B0%5D=0&votelst%5B0%5D=30000&votekd%5B0%5D=100" {
		t.Error("Link should be a known deeplink fee")
	}

}

func Test_ListParticipationKeys(t *testing.T) {
	ctx := context.Background()
	client, err := api.NewClientWithResponses("https://mainnet-api.4160.nodely.dev:443")
	if err != nil {
		t.Fatal(err)
	}

	_, err = GetKeys(ctx, client)

	// Expect unauthorized for Urtho servers
	if err == nil {
		t.Fatal(err)
	}

	// Setup elevated client
	tClient := test.GetClient(false)

	keys, err := GetKeys(ctx, tClient)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(keys)
}

func Test_ReadParticipationKey(t *testing.T) {
	ctx := context.Background()
	client, err := api.NewClientWithResponses("https://mainnet-api.4160.nodely.dev:443")
	if err != nil {
		t.Fatal(err)
	}

	_, err = GetKey(ctx, client, "unknown")

	// Expect unauthorized for Urtho servers
	if err == nil {
		t.Fatal(err)
	}

	tClient := test.GetClient(false)

	keys, err := GetKeys(ctx, tClient)
	if err != nil {
		t.Fatal(err)
	}
	if keys == nil {
		t.Fatal(err)
	}

	_, err = GetKey(ctx, tClient, (*keys)[0].Id)

	if err != nil {
		t.Fatal(err)
	}

}

func Test_GenerateParticipationKey(t *testing.T) {
	ctx := context.Background()

	// Create Client
	client, err := api.NewClientWithResponses("https://mainnet-api.4160.nodely.dev:443")
	if err != nil {
		t.Fatal(err)
	}
	// Generate error
	_, err = GenerateKeys(ctx, client, "", nil)
	if err == nil {
		t.Fatal(err)
	}

	// Setup test client
	tClient := test.GetClient(false)

	params := api.GenerateParticipationKeysParams{
		Dilution: nil,
		First:    0,
		Last:     30,
	}

	// This returns nothing and sucks
	key, err := GenerateKeys(ctx, tClient, "ABC", &params)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(key)
}

func Test_DeleteParticipationKey(t *testing.T) {
	ctx := context.Background()

	client := test.GetClient(false)
	params := api.GenerateParticipationKeysParams{
		Dilution: nil,
		First:    0,
		Last:     30000,
	}
	key, err := GenerateKeys(ctx, client, "ABC", &params)
	if err != nil {
		t.Fatal(err)
	}

	err = Delete(ctx, client, key.Id)
	if err != nil {
		t.Fatal(err)
	}
}
func Test_RemovePartKeyByID(t *testing.T) {
	// Test case: Remove an existing key
	t.Run("Remove existing key", func(t *testing.T) {
		keys := []api.ParticipationKey{
			{Id: "key1"},
			{Id: "key2"},
			{Id: "key3"},
		}
		expectedKeys := []api.ParticipationKey{
			{Id: "key1"},
			{Id: "key3"},
		}
		RemovePartKeyByID(&keys, "key2")
		if len(keys) != len(expectedKeys) {
			t.Fatalf("expected %d keys, got %d", len(expectedKeys), len(keys))
		}
		for i, key := range keys {
			if key.Id != expectedKeys[i].Id {
				t.Fatalf("expected key ID %s, got %s", expectedKeys[i].Id, key.Id)
			}
		}
	})

	// Test case: Remove a non-existing key
	t.Run("Remove non-existing key", func(t *testing.T) {
		keys := []api.ParticipationKey{
			{Id: "key1"},
			{Id: "key2"},
			{Id: "key3"},
		}
		expectedKeys := []api.ParticipationKey{
			{Id: "key1"},
			{Id: "key2"},
			{Id: "key3"},
		}
		RemovePartKeyByID(&keys, "key4")
		if len(keys) != len(expectedKeys) {
			t.Fatalf("expected %d keys, got %d", len(expectedKeys), len(keys))
		}
		for i, key := range keys {
			if key.Id != expectedKeys[i].Id {
				t.Fatalf("expected key ID %s, got %s", expectedKeys[i].Id, key.Id)
			}
		}
	})

	// Test case: Remove a key from an empty list
	t.Run("Remove key from empty list", func(t *testing.T) {
		keys := []api.ParticipationKey{}
		RemovePartKeyByID(&keys, "key1")
		if len(keys) != 0 {
			t.Fatalf("expected 0 keys, got %d", len(keys))
		}
	})
}
