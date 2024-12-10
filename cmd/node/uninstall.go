package node

import (
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:               "uninstall",
	Short:             "Uninstall Algorand node (Algod)",
	Long:              "Uninstall Algorand node (Algod) and other binaries on your system installed by this tool.",
	SilenceUsage:      true,
	PersistentPreRunE: NeedsToBeStopped,
	RunE: func(cmd *cobra.Command, args []string) error {
		return algod.Uninstall()
	},
}
