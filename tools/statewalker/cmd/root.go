// Package cmd contains the cobra command controllers for the statewalker CLI,
// a test harness that provisions Algorand networks and walks algod into
// specific states so nodekit behavior can be reviewed end to end.
package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// PrivateTemplate is the embedded goal network template, injected by main.
var PrivateTemplate []byte

// networkDir is the root directory holding the provisioned network(s).
var networkDir string

// rootCmd is the top level command of the statewalker CLI.
var rootCmd = &cobra.Command{
	Use:   "statewalker",
	Short: "Walk algod into specific states for nodekit e2e reviews",
	Long: `statewalker provisions Algorand test networks (private or algokit localnet)
and drives algod into reproducible states (partkey expirations, syncing,
fast-catchup, upgrade voting) so nodekit can be reviewed against them.`,
	SilenceUsage: true,
}

// DefaultNetworkDir returns the default root directory for statewalker networks.
func DefaultNetworkDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "statewalker", "net")
	}
	return filepath.Join(home, ".statewalker", "net")
}

// Execute runs the statewalker root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&networkDir, "network-dir", DefaultNetworkDir(), "root directory of the provisioned network")
	rootCmd.AddCommand(networkCmd)
}
