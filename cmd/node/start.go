package node

import (
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:               "start",
	Short:             "Start Algod",
	Long:              "Start Algod on your system (the one on your PATH).",
	SilenceUsage:      true,
	PersistentPreRunE: NeedsToBeStopped,
	RunE: func(cmd *cobra.Command, args []string) error {
		return algod.Start()
	},
}
