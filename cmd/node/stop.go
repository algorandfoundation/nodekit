package node

import (
	"fmt"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/algorandfoundation/algorun-tui/ui/style"
	"github.com/charmbracelet/log"
	"time"

	"github.com/spf13/cobra"
)

const StopTimeout = 5 * time.Second
const StopSuccessMsg = "Algod stopped successfully"
const StopFailureMsg = "failed to stop Algod"

var stopCmd = &cobra.Command{
	Use:              "stop",
	Short:            "Stop Algod",
	Long:             "Stop the Algod process on your system.",
	SilenceUsage:     true,
	PersistentPreRun: NeedsToBeRunning,
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Info(style.Green.Render("Stopping Algod 😢"))
		// Warn user for prompt
		log.Warn(style.Yellow.Render(SudoWarningMsg))

		err := algod.Stop()
		if err != nil {
			return fmt.Errorf(StopFailureMsg)
		}
		time.Sleep(StopTimeout)

		if algod.IsRunning() {
			return fmt.Errorf(StopFailureMsg)
		}

		log.Info(style.Green.Render("Algorand stopped successfully 🎉"))
		return nil
	},
}

func init() {
	stopCmd.Flags().BoolVarP(&force, "force", "f", false, style.Yellow.Render("forcefully stop the node"))
}
