package algod

import (
	"context"
	"github.com/algorandfoundation/nodekit/api"
	"github.com/algorandfoundation/nodekit/internal/test/mock"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
	"testing"
	"time"
)

func Test_StateModel(t *testing.T) {
	t.Skip()
	// Setup elevated client
	apiToken, err := securityprovider.NewSecurityProviderApiKey("header", "X-Algo-API-Token", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	if err != nil {
		t.Fatal(err)
	}
	client, err := api.NewClientWithResponses("http://localhost:8080", api.WithRequestEditorFn(apiToken.Intercept))
	httpPkg := new(api.HttpPkg)
	state := StateModel{
		Watching: true,
		Status: Status{
			LastRound:   1337,
			NeedsUpdate: true,
			State:       SyncingState,
			Client:      client,
			HttpPkg:     httpPkg,
		},
		Metrics: Metrics{
			RoundTime: 0,
			TX:        0,
			RX:        0,
			TPS:       0,
			Client:    client,
			HttpPkg:   httpPkg,
		},
		Client:  client,
		Context: context.Background(),
	}
	count := 0
	go state.Watch(func(model *StateModel, err error) {
		if err != nil || model == nil {
			t.Error("Failed")
			return
		}
		count++
	}, context.Background(), new(mock.Clock))
	time.Sleep(5 * time.Second)
	// Stop the watcher
	state.Stop()
	if count == 0 {
		t.Fatal("Did not receive any updates")
	}
	if state.Status.LastRound <= 0 {
		t.Fatal("LastRound is stale")
	}
	t.Log(
		"Watching: ", state.Watching,
		"LastRound: ", state.Status.LastRound,
		"NeedsUpdate: ", state.Status.NeedsUpdate,
		"State: ", state.Status.State,
		"RoundTime: ", state.Metrics.RoundTime,
		"RX: ", state.Metrics.RX,
		"TX: ", state.Metrics.TX,
	)
}
