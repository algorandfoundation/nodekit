package internal

import (
	"context"
	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"testing"
)

func Test_StatusModel(t *testing.T) {
	m := StatusModel{LastRound: 0}
	if m.String() != "Last round: 0" {
		t.Fatal("expected \"Last round: 0\", got ", m.String())
	}
	algodClient, err := algod.MakeClient(
		"https://mainnet-api.4160.nodely.dev",
		"",
	)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	go func() {
		err := m.Watch(ctx, algodClient)
		if err != nil {
			t.Error(err)
			return
		}
	}()
}
