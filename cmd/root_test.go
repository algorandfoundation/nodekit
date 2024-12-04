package cmd

import (
	"github.com/spf13/viper"
	"os"
	"testing"
)

func clearViper() {
	viper.Set("algod-token", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	viper.Set("algod-endpoint", "http://localhost:8080")
	viper.Set("ALGORAND_DATA", "")
}

// Test the stub root command
func Test_ExecuteRootCommand(t *testing.T) {
	clearViper()

	// Execute
	err := rootCmd.Execute()
	// Should always fail due to no TTY
	if err == nil {
		t.Fatal(err)
	}

	t.Run("InitConfig", func(t *testing.T) {
		cwd, _ := os.Getwd()
		viper.Set("algod-token", "")
		viper.Set("algod-endpoint", "")
		t.Setenv("ALGORAND_DATA", cwd+"/testdata/Test_InitConfig")

		initConfig()
		algod := viper.Get("algod-endpoint")
		if algod == "" {
			t.Fatal("Invalid Algod")
		}
		if algod != "http://127.0.0.1:8080" {
			t.Fatal("Invalid Algod")
		}
		clearViper()
	})

	t.Run("InitConfigWithoutEndpoint", func(t *testing.T) {
		cwd, _ := os.Getwd()
		viper.Set("algod-token", "")
		viper.Set("algod-endpoint", "")
		t.Setenv("ALGORAND_DATA", cwd+"/testdata/Test_InitConfigWithoutEndpoint")

		initConfig()
		algod := viper.Get("algod-endpoint")
		if algod == "" {
			t.Fatal("Invalid Algod")
		}
		if algod != "http://127.0.0.1:8080" {
			t.Fatal("Invalid Algod")
		}
		clearViper()
	})

	t.Run("InitConfigWithAddress", func(t *testing.T) {
		cwd, _ := os.Getwd()
		viper.Set("algod-token", "")
		viper.Set("algod-endpoint", "")
		t.Setenv("ALGORAND_DATA", cwd+"/testdata/Test_InitConfigWithAddress")

		initConfig()
		algod := viper.Get("algod-endpoint")
		if algod == "" {
			t.Fatal("Invalid Algod")
		}
		if algod != "http://255.255.255.255:8080" {
			t.Fatal("Invalid Algod")
		}
		clearViper()
	})

	t.Run("InitConfigWithAddressAndDefaultPort", func(t *testing.T) {
		cwd, _ := os.Getwd()
		viper.Set("algod-token", "")
		viper.Set("algod-endpoint", "")
		t.Setenv("ALGORAND_DATA", cwd+"/testdata/Test_InitConfigWithAddressAndDefaultPort")

		initConfig()
		algod := viper.Get("algod-endpoint")
		if algod == "" {
			t.Fatal("Invalid Algod")
		}
		if algod != "http://255.255.255.255:8080" {
			t.Fatal("Invalid Algod")
		}
		clearViper()
	})
}
