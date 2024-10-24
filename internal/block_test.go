package internal

import (
	"context"
	"github.com/algorandfoundation/hack-tui/api"
	"testing"
	"time"
)

func Test_GetBlockMetrics(t *testing.T) {
	window := 1000000

	expectedAvg := time.Duration(2856041000)

	client, err := api.NewClientWithResponses("https://mainnet-api.4160.nodely.dev:443")

	metrics, err := GetBlockMetrics(context.Background(), client, uint64(42000000), window)
	if err != nil {
		t.Fatal(err)
	}

	if metrics.AvgTime != expectedAvg {
		t.Fatal("expected time to be", expectedAvg, "got", metrics.AvgTime)
	}

	expectedTPS := 25.318608871511294

	if metrics.TPS != expectedTPS {
		t.Fatal("expected tps to be", expectedTPS, "got", metrics.TPS)
	}
}
