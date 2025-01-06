package cmd

import (
	"github.com/algorandfoundation/algorun-tui/cmd/utils/explanations"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/algorandfoundation/algorun-tui/ui/style"
	"github.com/charmbracelet/log"
	"time"

	"github.com/spf13/cobra"
)

// StopTimeout defines the duration to wait after attempting to stop the Algod process to ensure it has fully shut down.
const StopTimeout = 5 * time.Second

const StoppingAlgodMsg = "Stopping Algod 😢"

// StopSuccessMsg is a constant string message indicating that Algod has been stopped successfully.
const StopSuccessMsg = "Algorand stopped successfully 🎉"

// StopFailureMsg is a constant string used as an error message when the Algod process fails to stop.
const StopFailureMsg = "failed to stop Algod"

var stopCmd = &cobra.Command{
	Use:              "stop",
	Short:            "Stop Algod",
	Long:             "Stop the Algod process on your system.",
	SilenceUsage:     true,
	PersistentPreRun: NeedsToBeRunning,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info(style.Green.Render(StoppingAlgodMsg))
		// Warn user for prompt
		log.Warn(style.Yellow.Render(explanations.SudoWarningMsg))

		err := algod.Stop()
		if err != nil {
			log.Fatal(StopFailureMsg)
		}
		time.Sleep(StopTimeout)

		if algod.IsRunning() {
			log.Fatal(StopFailureMsg)
		}

		log.Info(style.Green.Render(StopSuccessMsg))
	},
}

func init() {
	stopCmd.Flags().BoolVarP(&force, "force", "f", false, style.Yellow.Render("forcefully stop the node"))
}
