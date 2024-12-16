package node

import (
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/algorandfoundation/algorun-tui/ui/style"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

const SudoWarningMsg = "(You may be prompted for your password)"
const PermissionErrorMsg = "this command must be run with super-user privileges (sudo)"
const NotInstalledErrorMsg = "algod is not installed. please run the *node install* command"
const RunningErrorMsg = "algod is running, please run the *node stop* command"
const NotRunningErrorMsg = "algod is not running"

var (
	force bool = false
)
var Cmd = &cobra.Command{
	Use:   "node",
	Short: "Node Management",
	Long:  style.Purple(style.BANNER) + "\n" + style.LightBlue("Manage your Algorand node"),
}

func NeedsToBeRunning(cmd *cobra.Command, args []string) {
	if force {
		return
	}
	if !algod.IsInstalled() {
		log.Fatal(NotInstalledErrorMsg)
	}
	if !algod.IsRunning() {
		log.Fatal(NotRunningErrorMsg)
	}
}

func NeedsToBeStopped(cmd *cobra.Command, args []string) {
	if force {
		return
	}
	if !algod.IsInstalled() {
		log.Fatal(NotInstalledErrorMsg)
	}
	if algod.IsRunning() {
		log.Fatal(RunningErrorMsg)
	}
}

func init() {
	Cmd.AddCommand(installCmd)
	Cmd.AddCommand(startCmd)
	Cmd.AddCommand(stopCmd)
	Cmd.AddCommand(uninstallCmd)
	Cmd.AddCommand(upgradeCmd)
	Cmd.AddCommand(syncCmd)
	Cmd.AddCommand(debugCmd)
}
