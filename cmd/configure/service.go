package configure

import (
	"errors"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/algorandfoundation/algorun-tui/internal/system"
	"github.com/algorandfoundation/algorun-tui/ui/style"
	"github.com/spf13/cobra"
)

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Configure the node service",
	Long:  style.Purple(style.BANNER) + "\n" + style.LightBlue("Configure the service that runs the node."),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if !system.IsSudo() {
			return errors.New(
				"you need to be root to run this command. Please run this command with sudo")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Combine this with algod.UpdateService and algod.SetNetwork
		return algod.EnsureService()
	},
}
