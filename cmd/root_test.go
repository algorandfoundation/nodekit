package cmd

import (
	"errors"
	"github.com/algorandfoundation/algorun-tui/cmd/utils"
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

	t.Run("Invalid algod-endpoint", func(t *testing.T) {
		viper.Set("algod-endpoint", "")
		err := rootCmd.Execute()
		if err == nil {
			t.Error("No error for invalid algod-endpoint")
		}
		clearViper()
	})
	t.Run("Invalid algod-token", func(t *testing.T) {
		viper.Set("algod-endpoint", "http://localhost:8080")
		viper.Set("algod-token", "")
		err := rootCmd.Execute()
		if err == nil {
			t.Error("No error for invalid algod-endpoint")
		}
	})
	t.Run("InitConfig", func(t *testing.T) {
		cwd, _ := os.Getwd()
		viper.Set("algod-token", "")
		viper.Set("algod-endpoint", "")
		t.Setenv("ALGORAND_DATA", cwd+"/testdata/Test_InitConfig")

		_ = utils.InitConfig()
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

		_ = utils.InitConfig()
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

		_ = utils.InitConfig()
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

		_ = utils.InitConfig()
		algod := viper.Get("algod-endpoint")
		if algod == "" {
			t.Fatal("Invalid Algod")
		}
		if algod != "http://255.255.255.255:8080" {
			t.Fatal("Invalid Algod")
		}
		clearViper()
	})

	t.Run("check error", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}
		}()
		check(errors.New("test"))
	})
}
