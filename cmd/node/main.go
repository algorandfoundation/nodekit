package node

import (
	"github.com/algorandfoundation/hack-tui/ui/style"
	"github.com/spf13/cobra"
)

var NodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Algod installation",
	Long:  style.Purple(style.BANNER) + "\n" + style.LightBlue("View the node status"),
}

func init() {
	NodeCmd.AddCommand(configureCmd)
	NodeCmd.AddCommand(installCmd)
	NodeCmd.AddCommand(startCmd)
	NodeCmd.AddCommand(stopCmd)
	NodeCmd.AddCommand(uninstallCmd)
	NodeCmd.AddCommand(upgradeCmd)
}
