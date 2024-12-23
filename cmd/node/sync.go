package node

import (
	"context"
	"github.com/algorandfoundation/algorun-tui/api"
	"github.com/algorandfoundation/algorun-tui/cmd/utils"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/algorandfoundation/algorun-tui/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// dataDir path to the algorand data folder
var dataDir string = ""

// defaultLag represents the default minimum catchup delay in milliseconds for the Fast Catchup process.
var defaultLag int = 30_000

var syncCmdShortTxt = "Quickly catch up your node to the latest block."
var syncCmdLongTxt = lipgloss.JoinVertical(
	lipgloss.Left,
	style.Purple(style.BANNER),
	"",
	style.Bold(syncCmdShortTxt),
	"",
	style.BoldUnderline("Overview:"),
	"Fetch the latest catchpoint and use Fast-Catchup to check if the node is caught up to the latest block.",
	"",
	style.Yellow.Render("Note: Not all networks support Fast-Catchup."),
)

// syncCmd is a Cobra command used to check the node's sync status and initiate a fast catchup when necessary.
var syncCmd = utils.WithAlgodFlags(&cobra.Command{
	Use:          "sync",
	Short:        syncCmdShortTxt,
	Long:         syncCmdLongTxt,
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		// Create Clients
		ctx := context.Background()
		httpPkg := new(api.HttpPkg)
		client, err := algod.GetClient(dataDir)
		cobra.CheckErr(err)

		// Fetch Status from Node
		status, response, err := algod.NewStatus(ctx, client, httpPkg)
		utils.WithInvalidResponsesExplanations(err, response, cmd.UsageString())
		if status.State == algod.FastCatchupState {
			log.Fatal(style.Red.Render("Node is currently catching up. Use --abort to cancel."))
		}

		// Get the Latest Catchpoint
		catchpoint, _, err := algod.GetLatestCatchpoint(httpPkg, status.Network)
		if err != nil {
			log.Fatal(err)
		}
		log.Info(style.Green.Render("Latest Catchpoint: " + catchpoint))

		// Submit the Catchpoint to the Algod Node, using the StartCatchupParams to skip
		res, _, err := algod.StartCatchup(ctx, client, catchpoint, &api.StartCatchupParams{Min: &defaultLag})
		if err != nil {
			log.Fatal(err)
		}

		log.Info(style.Green.Render(res))
	},
}, &dataDir)
