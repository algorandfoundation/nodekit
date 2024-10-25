package internal

import (
	"context"
	"fmt"
	"testing"

	"github.com/algorandfoundation/hack-tui/api"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
)

func Test_ListParticipationKeys(t *testing.T) {
	ctx := context.Background()
	client, err := api.NewClientWithResponses("https://mainnet-api.4160.nodely.dev:443")
	if err != nil {
		t.Fatal(err)
	}

	_, err = GetPartKeys(ctx, client)

	// Expect unauthorized for Urtho servers
	if err == nil {
		t.Fatal(err)
	}

	// Setup elevated client
	apiToken, err := securityprovider.NewSecurityProviderApiKey("header", "X-Algo-API-Token", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatal(err)
	}
	client, err = api.NewClientWithResponses("http://localhost:8080", api.WithRequestEditorFn(apiToken.Intercept))
	if err != nil {
		t.Fatal(err)
	}

	keys, err := GetPartKeys(ctx, client)
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

	_, err = ReadPartKey(ctx, client, "unknown")

	// Expect unauthorized for Urtho servers
	if err == nil {
		t.Fatal(err)
	}

	// Setup elevated client
	apiToken, err := securityprovider.NewSecurityProviderApiKey("header", "X-Algo-API-Token", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatal(err)
	}
	client, err = api.NewClientWithResponses("http://localhost:8080", api.WithRequestEditorFn(apiToken.Intercept))
	if err != nil {
		t.Fatal(err)
	}

	keys, err := GetPartKeys(ctx, client)
	if err != nil {
		t.Fatal(err)
	}
	if keys == nil {
		t.Fatal(err)
	}

	_, err = ReadPartKey(ctx, client, (*keys)[0].Id)

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
	_, err = GenerateKeyPair(ctx, client, "", nil)
	if err == nil {
		t.Fatal(err)
	}

	// Setup elevated client
	apiToken, err := securityprovider.NewSecurityProviderApiKey("header", "X-Algo-API-Token", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatal(err)
	}
	client, err = api.NewClientWithResponses("http://localhost:8080", api.WithRequestEditorFn(apiToken.Intercept))
	if err != nil {
		t.Fatal(err)
	}

	params := api.GenerateParticipationKeysParams{
		Dilution: nil,
		First:    0,
		Last:     10000,
	}

	// This returns nothing and sucks
	key, err := GenerateKeyPair(ctx, client, "QNZ7GONNHTNXFW56Y24CNJQEMYKZKKI566ASNSWPD24VSGKJWHGO6QOP7U", &params)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(key)
}

func Test_DeleteParticipationKey(t *testing.T) {
	ctx := context.Background()
	// Setup elevated client
	apiToken, err := securityprovider.NewSecurityProviderApiKey("header", "X-Algo-API-Token", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatal(err)
	}
	client, err := api.NewClientWithResponses("http://localhost:8080", api.WithRequestEditorFn(apiToken.Intercept))
	if err != nil {
		t.Fatal(err)
	}
	params := api.GenerateParticipationKeysParams{
		Dilution: nil,
		First:    0,
		Last:     10000,
	}
	key, err := GenerateKeyPair(ctx, client, "QNZ7GONNHTNXFW56Y24CNJQEMYKZKKI566ASNSWPD24VSGKJWHGO6QOP7U", &params)
	if err != nil {
		t.Fatal(err)
	}

	err = DeletePartKey(ctx, client, key.Id)
	if err != nil {
		t.Fatal(err)
	}
}
