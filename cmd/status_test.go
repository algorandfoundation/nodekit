package cmd

import (
	"context"
	"github.com/spf13/viper"
	"testing"
)

func Test_ExecuteInvalidStatusCommand(t *testing.T) {
	viper.Set("server", "")
	err := statusCmd.RunE(nil, nil)
	if err == nil {
		t.Error("Must fail when server is missing")
	}
}

// Test the Status Command
func Test_ExecuteStatusCommand(t *testing.T) {
	// Smoke Test Command
	viper.Set("server", "https://mainnet-api.4160.nodely.dev:443")
	go func() {
		err := statusCmd.RunE(nil, []string{"--server", "https://mainnet-api.4160.nodely.dev:443"})
		if err != nil {

		}
	}()
	context.Background().Done()
}
