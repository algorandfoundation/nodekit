package cmd

import (
	cmdutils "github.com/algorandfoundation/nodekit/cmd/utils"
	"github.com/algorandfoundation/nodekit/cmd/utils/explanations"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// StartingAlgodMsg is a constant string message indicating that Algod is being started.
const StartingAlgodMsg = "Starting Algod 🚀"

// StartSuccessMsg is a constant string message indicating that Algod has been started successfully.
const StartSuccessMsg = "Algorand started successfully 🎉"

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
var startCmd = cmdutils.WithAlgodFlags(&cobra.Command{
	Use:              "start",
	Short:            startShort,
	Long:             startLong,
	SilenceUsage:     true,
	PersistentPreRun: NeedsToBeStopped,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info(style.Green.Render(StartingAlgodMsg))
		// Warn user for prompt
		log.Warn(style.Yellow.Render(explanations.SudoWarningMsg))
		err := algod.Start()
		if err != nil {
			log.Fatal(err)
		}
		log.Info(style.Green.Render(StartSuccessMsg))
	},
}, &algodData)

// init initializes the `force` flag for the `start` command, allowing the node to start forcefully when specified.
func init() {
	startCmd.Flags().BoolVarP(&force, "force", "f", false, style.Yellow.Render("forcefully start the node"))
}
