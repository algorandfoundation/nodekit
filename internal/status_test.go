package internal

import (
	"context"
	"github.com/algorandfoundation/hack-tui/api"
	"testing"
)

func Test_StatusModel(t *testing.T) {
	m := StatusModel{LastRound: 0}
	if m.String() != "Last round: 0" {
		t.Fatal("expected \"Last round: 0\", got ", m.String())
	}

	client, err := api.NewClientWithResponses("https://mainnet-api.4160.nodely.dev:443")
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	go func() {
		err := m.Watch(ctx, client)
		if err != nil {
			t.Error(err)
			return
		}
	}()
}
