package node

import (
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/algorandfoundation/algorun-tui/ui/style"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:              "start",
	Short:            "Start Algod",
	Long:             "Start Algod on your system (the one on your PATH).",
	SilenceUsage:     true,
	PersistentPreRun: NeedsToBeStopped,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info(style.Green.Render("Starting Algod 🚀"))
		// Warn user for prompt
		log.Warn(style.Yellow.Render(SudoWarningMsg))
		err := algod.Start()
		if err != nil {
			log.Fatal(err)
		}
		log.Info(style.Green.Render("Algorand started successfully 🎉"))
	},
}

func init() {
	startCmd.Flags().BoolVarP(&force, "force", "f", false, style.Yellow.Render("forcefully start the node"))
}
