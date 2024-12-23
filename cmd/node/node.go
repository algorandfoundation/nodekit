package node

import (
	"github.com/algorandfoundation/algorun-tui/cmd/node/catchup"
	"github.com/algorandfoundation/algorun-tui/cmd/node/configure"
	"github.com/algorandfoundation/algorun-tui/cmd/utils/explanations"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/algorandfoundation/algorun-tui/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// SudoWarningMsg is a constant string displayed to warn users that they may be prompted for their password during execution.
const SudoWarningMsg = "(You may be prompted for your password)"

// PermissionErrorMsg is a constant string that indicates a command requires super-user privileges (sudo) to be executed.
const PermissionErrorMsg = "this command must be run with super-user privileges (sudo)"

// NotInstalledErrorMsg is the error message displayed when the algod software is not installed on the system.
const NotInstalledErrorMsg = "algod is not installed. please run the *node install* command"

// RunningErrorMsg represents the error message displayed when algod is running and needs to be stopped before proceeding.
const RunningErrorMsg = "algod is running, please run the *node stop* command"

// NotRunningErrorMsg is the error message displayed when the algod service is not currently running on the system.
const NotRunningErrorMsg = "algod is not running"

// force indicates whether actions should be performed forcefully, bypassing checks or confirmations.
var force bool = false

var short = "Manage your Algorand node using the CLI."
var long = lipgloss.JoinVertical(
	lipgloss.Left,
	style.Purple(style.BANNER),
	"",
	style.Bold(short),
	"",
	style.BoldUnderline("Overview:"),
	"A collection of commands for installing, configuring, starting, stopping, and upgrading your node.",
	"",
	style.Yellow.Render(explanations.ExperimentalWarning),
)

// Cmd represents the root command for managing an Algorand node, providing subcommands for installation, control, and upgrades.
var Cmd = &cobra.Command{
	Use:   "node",
	Short: short,
	Long:  long,
}

// NeedsToBeRunning ensures the Algod software is installed and running before executing the associated Cobra command.
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

// NeedsToBeStopped ensures the operation halts if Algod is not installed or is currently running, unless forced.
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

// init initializes the root command by adding subcommands for managing the Algorand node, such as install, start, stop, etc.
func init() {
	Cmd.AddCommand(installCmd)
	Cmd.AddCommand(startCmd)
	Cmd.AddCommand(stopCmd)
	Cmd.AddCommand(uninstallCmd)
	Cmd.AddCommand(upgradeCmd)
	Cmd.AddCommand(syncCmd)
	Cmd.AddCommand(debugCmd)
	Cmd.AddCommand(bootstrapCmd)
	Cmd.AddCommand(catchup.Cmd)
	Cmd.AddCommand(configure.Cmd)
}
