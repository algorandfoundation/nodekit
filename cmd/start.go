package cmd

import (
	"github.com/algorandfoundation/nodekit/cmd/utils"
	"github.com/algorandfoundation/nodekit/cmd/utils/explanations"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var startShort = "Start the node daemon"

var startLong = lipgloss.JoinVertical(
	lipgloss.Left,
	style.Purple(style.BANNER),
	"",
	style.Bold(startShort),
	"",
	style.BoldUnderline("Overview:"),
	"Start the Algorand daemon on your local machine if it is not already running. Optionally, the daemon can be forcefully started.",
	"",
	style.Yellow.Render("This requires the daemon to be installed on your system."),
)

// startCmd is a Cobra command used to start the Algod service on the system, ensuring necessary checks are performed beforehand.
var startCmd = utils.WithAlgodFlags(&cobra.Command{
	Use:              "start",
	Short:            startShort,
	Long:             startLong,
	SilenceUsage:     true,
	PersistentPreRun: NeedsToBeStopped,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info(style.Green.Render("Starting Algod 🚀"))
		// Warn user for prompt
		log.Warn(style.Yellow.Render(explanations.SudoWarningMsg))
		err := algod.Start(dataDir)
		if err != nil {
			log.Fatal(err)
		}
		log.Info(style.Green.Render("Algorand started successfully 🎉"))
	},
}, &dataDir)

// init initializes the `force` flag for the `start` command, allowing the node to start forcefully when specified.
func init() {
	startCmd.Flags().BoolVarP(&force, "force", "f", false, style.Yellow.Render("forcefully start the node"))
}
