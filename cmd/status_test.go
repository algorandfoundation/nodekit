package cmd

import (
	"context"
	"github.com/spf13/viper"
	"testing"
)

func Test_ExecuteInvalidStatusCommand(t *testing.T) {
	viper.Set("algod-endpoint", "")
	err := statusCmd.RunE(nil, nil)
	if err == nil {
		t.Error("Must fail when algod-endpoint is missing")
	}
}

// Test the Status Command
func Test_ExecuteStatusCommand(t *testing.T) {
	// Smoke Test Command
	viper.Set("algod-endpoint", "https://mainnet-api.4160.nodely.dev:443")
	go func() {
		err := statusCmd.RunE(nil, []string{"--algod-endpoint", "https://mainnet-api.4160.nodely.dev:443"})
		if err != nil {

		}
	}()
	context.Background().Done()
}
