package participation

import (
	"bytes"
	"context"
	"fmt"
	"github.com/algorandfoundation/algorun-tui/api"
	"github.com/algorandfoundation/algorun-tui/internal/test"
	"github.com/algorandfoundation/algorun-tui/internal/test/mock"
	"io"
	"net/http"
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

	_, _, err = GetList(ctx, client)

	// Expect unauthorized for Urtho servers
	if err == nil {
		t.Fatal(err)
	}

	// Setup elevated client
	tClient := test.GetClient(false)

	keys, _, err := GetList(ctx, tClient)
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

	_, _, err = GetKey(ctx, client, "unknown")

	// Expect unauthorized for Urtho servers
	if err == nil {
		t.Fatal(err)
	}

	tClient := test.GetClient(false)

	keys, _, err := GetList(ctx, tClient)
	if err != nil {
		t.Fatal(err)
	}
	if keys == nil {
		t.Fatal(err)
	}

	_, _, err = GetKey(ctx, tClient, (keys)[0].Id)

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
		keys := List{
			{Id: "key1"},
			{Id: "key2"},
			{Id: "key3"},
		}
		expectedKeys := List{
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
		keys := List{
			{Id: "key1"},
			{Id: "key2"},
			{Id: "key3"},
		}
		expectedKeys := List{
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
		keys := List{}
		RemovePartKeyByID(&keys, "key1")
		if len(keys) != 0 {
			t.Fatalf("expected 0 keys, got %d", len(keys))
		}
	})
}

var onlineShortLinkResponseStr = `{
    "id": "WKIPKTWIFZQJ2"
  }`

type testOnlineShortner struct {
	api.HttpPkgInterface
}

// TODO: toggle between Unit/Integration tests
func (testOnlineShortner) Post(url string, bodyType string, body io.Reader) (resp *http.Response, err error) {
	responseBody := io.NopCloser(bytes.NewReader([]byte(onlineShortLinkResponseStr)))
	return &http.Response{
		Status:           "",
		StatusCode:       0,
		Proto:            "",
		ProtoMajor:       0,
		ProtoMinor:       0,
		Header:           nil,
		Body:             responseBody,
		ContentLength:    0,
		TransferEncoding: nil,
		Close:            false,
		Uncompressed:     false,
		Trailer:          nil,
		Request:          nil,
		TLS:              nil,
	}, nil
}
func Test_ToOnlineShortLink(t *testing.T) {
	link, err := GetOnlineShortLink(new(testOnlineShortner), OnlineShortLinkBody{
		Account:          "JPEGRZ6G4IBZCOC7UV6QZWJ6TENNKRIPENUJTLG5K7PKIKMVTJHUGERARE",
		VoteKeyB64:       "WWHePYtNZ2T3sHkqdd/38EvoFWrnIKPrTo6xN/4T1l4=",
		SelectionKeyB64:  "e4kBLu7zXOorjLVzJHOiAn+IhOBsYBCqqHKaJCiCdJs=",
		StateProofKeyB64: "1GdNPOck+t6yXvuXxrDEPKqgi4I2sTaNugV1kd5ksUW2G1U6x1FT0WR3aT3ZSSmbYoDt3cVrB3vIPJA8GkqSYg==",
		VoteFirstValid:   3118965,
		VoteLastValid:    3148965,
		KeyDilution:      995,
		Network:          "mainnet",
	})
	if err != nil {
		t.Error(err)
	}
	if link.Id != "WKIPKTWIFZQJ2" {
		t.Error("Link should be a known hash")
	}
}

var offlineShortLinkResponseStr = `{
    "id": "D3O3GEG2UD2GW"
  }`

type testOfflineShortner struct {
	api.HttpPkgInterface
}

// TODO: toggle between Unit/Integration tests
func (testOfflineShortner) Post(url string, bodyType string, body io.Reader) (resp *http.Response, err error) {
	responseBody := io.NopCloser(bytes.NewReader([]byte(offlineShortLinkResponseStr)))
	return &http.Response{
		Status:           "",
		StatusCode:       0,
		Proto:            "",
		ProtoMajor:       0,
		ProtoMinor:       0,
		Header:           nil,
		Body:             responseBody,
		ContentLength:    0,
		TransferEncoding: nil,
		Close:            false,
		Uncompressed:     false,
		Trailer:          nil,
		Request:          nil,
		TLS:              nil,
	}, nil
}
func Test_ToOfflineShortLink(t *testing.T) {
	link, err := GetOfflineShortLink(new(testOfflineShortner), OfflineShortLinkBody{
		Account: "JPEGRZ6G4IBZCOC7UV6QZWJ6TENNKRIPENUJTLG5K7PKIKMVTJHUGERARE",
		Network: "mainnet",
	})
	if err != nil {
		t.Error(err)
	}
	if link.Id != "D3O3GEG2UD2GW" {
		t.Error("Link should be a known hash")
	}
}
