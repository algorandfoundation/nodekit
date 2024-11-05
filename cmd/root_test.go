package cmd

import (
	"github.com/spf13/viper"
	"os"
	"testing"
)

// Test the stub root command
func Test_ExecuteRootCommand(t *testing.T) {
	viper.Set("token", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	viper.Set("server", "http://localhost:8080")

	// Execute
	err := rootCmd.Execute()
	// Should always fail due to no TTY
	if err == nil {
		t.Fatal(err)
	}
}

func Test_InitConfig(t *testing.T) {
	cwd, _ := os.Getwd()
	viper.Set("token", "")
	viper.Set("server", "")
	t.Setenv("ALGORAND_DATA", cwd+"/testdata/Test_InitConfig")

	initConfig()
	server := viper.Get("server")
	if server == "" {
		t.Fatal("Invalid Server")
	}
	if server != "http://127.0.0.1:8080" {
		t.Fatal("Invalid Server")
	}
}

func Test_InitConfigWithoutEndpoint(t *testing.T) {
	cwd, _ := os.Getwd()
	t.Setenv("ALGORAND_DATA", cwd+"/testdata/Test_InitConfigWithoutEndpoint")

	initConfig()
	server := viper.Get("server")
	if server == "" {
		t.Fatal("Invalid Server")
	}
	if server != "http://127.0.0.1:8080" {
		t.Fatal("Invalid Server")
	}
}

func Test_InitConfigWithAddress(t *testing.T) {
	cwd, _ := os.Getwd()
	t.Setenv("ALGORAND_DATA", cwd+"/testdata/Test_InitConfigWithAddress")

	initConfig()
	server := viper.Get("server")
	if server == "" {
		t.Fatal("Invalid Server")
	}
	if server != "http://255.255.255.255:8080" {
		t.Fatal("Invalid Server")
	}
}
