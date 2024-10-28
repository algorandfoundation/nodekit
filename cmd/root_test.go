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
	t.Setenv("ALGORAND_DATA", cwd+"/testdata")

	initConfig()
	server := viper.Get("server")
	if server == "" {
		t.Fatal("Invalid Server")
	}

}
