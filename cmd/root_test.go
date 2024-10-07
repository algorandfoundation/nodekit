package cmd

import (
	"bytes"
	"github.com/spf13/viper"
	"io"
	"os"
	"testing"
)

// Test the stub root command
func Test_ExecuteRootCommand(t *testing.T) {
	// Set output
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{"--server", "https://mainnet-api.4160.nodely.dev:443"})

	// Execute
	err := rootCmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

	// Read the buffer
	out, err := io.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}

	// Assert the command
	if string(out) != "INFO Arguments: Server: https://mainnet-api.4160.nodely.dev:443\n" {
		t.Fatalf("expected \"%s\" got \"%s\"", "hi", string(out))
	}
}

func Test_ExecuteRootCommand_NoArgs(t *testing.T) {
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
}

func Test_InitConfig(t *testing.T) {
	cwd, _ := os.Getwd()
	t.Setenv("ALGORAND_DATA", cwd+"/testdata")

	initConfig()
	server := viper.Get("server")
	if server == "" {
		t.Fatal("Wow")
	}

}
