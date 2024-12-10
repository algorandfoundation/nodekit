package node

import (
	"errors"
	"fmt"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/algorandfoundation/algorun-tui/ui/style"
	"github.com/spf13/cobra"
	"os"
	"runtime"
)

const PermissionErrorMsg = "this command must be run with super-user privileges (sudo)"
const NotInstalledErrorMsg = "algod is not installed. please run the *node install* command"
const RunningErrorMsg = "algod is running, please run the *node stop* command"
const NotRunningErrorMsg = "algod is not running"

var Cmd = &cobra.Command{
	Use:   "node",
	Short: "Node Management",
	Long:  style.Purple(style.BANNER) + "\n" + style.LightBlue("Manage your Algorand node"),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Check that we are calling with sudo on linux
		if os.Geteuid() != 0 && runtime.GOOS == "linux" {
			return errors.New(PermissionErrorMsg)
		}
		return nil
	},
}

func NeedsToBeRunning(cmd *cobra.Command, args []string) error {
	if !algod.IsInstalled() {
		return fmt.Errorf(NotInstalledErrorMsg)
	}
	if !algod.IsRunning() {
		return fmt.Errorf(NotRunningErrorMsg)
	}
	return nil
}

func NeedsToBeStopped(cmd *cobra.Command, args []string) error {
	if !algod.IsInstalled() {
		return fmt.Errorf(NotInstalledErrorMsg)
	}
	if algod.IsRunning() {
		return fmt.Errorf(RunningErrorMsg)
	}
	return nil
}

func init() {
	Cmd.AddCommand(installCmd)
	Cmd.AddCommand(startCmd)
	Cmd.AddCommand(stopCmd)
	Cmd.AddCommand(uninstallCmd)
	Cmd.AddCommand(upgradeCmd)
	Cmd.AddCommand(debugCmd)
}
