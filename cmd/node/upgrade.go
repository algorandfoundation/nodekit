package node

import (
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:              "upgrade",
	Short:            "Upgrade Algod",
	Long:             "Upgrade Algod (if installed with package manager).",
	SilenceUsage:     true,
	PersistentPreRun: NeedsToBeStopped,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Check Version from S3 against the local binary
		return algod.Update()
	},
}
