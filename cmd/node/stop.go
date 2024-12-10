package node

import (
	"fmt"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"time"

	"github.com/spf13/cobra"
)

const StopTimeout = 5 * time.Second
const StopSuccessMsg = "Algod stopped successfully"
const StopFailureMsg = "failed to stop Algod"

var stopCmd = &cobra.Command{
	Use:               "stop",
	Short:             "Stop Algod",
	Long:              "Stop the Algod process on your system.",
	SilenceUsage:      true,
	PersistentPreRunE: NeedsToBeRunning,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Stopping Algod...")
		err := algod.Stop()
		if err != nil {
			return fmt.Errorf(StopFailureMsg)
		}
		time.Sleep(StopTimeout)

		if algod.IsRunning() {
			return fmt.Errorf(StopFailureMsg)
		}

		fmt.Println(StopSuccessMsg)
		return nil
	},
}
