package cmd

import (
	"time"

	cmdutils "github.com/algorandfoundation/nodekit/cmd/utils"
	"github.com/algorandfoundation/nodekit/cmd/utils/explanations"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var restartShort = "Restart the node daemon"

var restartLong = lipgloss.JoinVertical(
	lipgloss.Left,
	style.Purple(style.BANNER),
	"",
	style.Bold(restartShort),
	"",
	style.BoldUnderline("Overview:"),
	"Stops the Algorand daemon on your local machine and starts it again. Optionally, the daemon can be forcefully restarted.",
	"",
	style.Yellow.Render("This requires the daemon to be installed and running on your system."),
)

// NeedsToBeRunningToRestart ensures Algod is installed and already running, pointing at the *start* command when it is not.
func NeedsToBeRunningToRestart(cmd *cobra.Command, args []string) {
	if force {
		return
	}
	if !algod.IsInstalled() {
		log.Fatal(explanations.NotInstalledErrorMsg)
	}
	if !algod.IsRunning(algodData) {
		log.Fatal(explanations.NotRunningStartErrorMsg)
	}
}

// restartCmd is a Cobra command that stops a running Algod service and starts it back up.
var restartCmd = cmdutils.WithAlgodFlags(&cobra.Command{
	Use:              "restart",
	Short:            restartShort,
	Long:             restartLong,
	SilenceUsage:     true,
	PersistentPreRun: NeedsToBeRunningToRestart,
	Run: func(cmd *cobra.Command, args []string) {
		// Warn user for prompt
		log.Warn(style.Yellow.Render(explanations.SudoWarningMsg))

		log.Info(style.Green.Render(StoppingAlgodMsg))
		err := algod.Stop()
		if err != nil {
			log.Fatal(StopFailureMsg)
		}
		time.Sleep(StopTimeout)

		if algod.IsRunning(algodData) {
			log.Fatal(StopFailureMsg)
		}
		log.Info(style.Green.Render(StopSuccessMsg))

		log.Info(style.Green.Render(StartingAlgodMsg))
		err = algod.Start()
		if err != nil {
			log.Fatal(err)
		}
		log.Info(style.Green.Render(StartSuccessMsg))
	},
}, &algodData)

// init initializes the `force` flag for the `restart` command, allowing the node to restart forcefully when specified.
func init() {
	restartCmd.Flags().BoolVarP(&force, "force", "f", false, style.Yellow.Render("forcefully restart the node"))
}
