package cmd

import (
	"github.com/algorandfoundation/algorun-tui/cmd/utils/explanations"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/algorandfoundation/algorun-tui/ui/style"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"os"
	"time"
)

// UpgradeMsg is a constant string used to indicate the start of the Algod upgrade process.
const UpgradeMsg = "Upgrading Algod"

// upgradeCmd is a Cobra command used to upgrade Algod, utilizing the OS-specific package manager if applicable.
var upgradeCmd = &cobra.Command{
	Use:              "upgrade",
	Short:            "Upgrade Algod",
	Long:             "Upgrade Algod (if installed with package manager).",
	SilenceUsage:     true,
	PersistentPreRun: NeedsToBeStopped,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: get expected version and check if update is required
		log.Info(style.Green.Render(UpgradeMsg))
		// Warn user for prompt
		log.Warn(style.Yellow.Render(explanations.SudoWarningMsg))
		// TODO: Check Version from S3 against the local binary
		err := algod.Update()
		if err != nil {
			log.Error(err)
		}

		time.Sleep(5 * time.Second)

		// If it's not running, start the daemon (can happen)
		if !algod.IsRunning() {
			err = algod.Start()
			if err != nil {
				log.Error(err)
				os.Exit(1)
			}
		}
	},
}
