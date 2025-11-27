package catchup

import (
	"context"
	"github.com/algorandfoundation/nodekit/api"
	"github.com/algorandfoundation/nodekit/cmd/utils"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// startCmdLong provides a detailed description and overview message for the 'start' command, including notes and caveats.
var startCmdLong = lipgloss.JoinVertical(
	lipgloss.Left,
	style.Purple(style.BANNER),
	"",
	style.Bold("Catchup the node to the latest catchpoint."),
	"",
	style.BoldUnderline("Overview:"),
	"Starting a catchup will sync the node to the latest catchpoint.",
	"Actual sync times may vary depending on the number of accounts, number of blocks and the network.",
	"",
	style.Yellow.Render("Note: Not all networks support Fast-Catchup."),
)

// startCmd is a Cobra command used to check the node's sync status and initiate a fast catchup when necessary.
var startCmd = utils.WithAlgodFlags(&cobra.Command{
	Use:          "start",
	Short:        "Get the latest catchpoint and start catching up.",
	Long:         startCmdLong,
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		httpPkg := new(api.HttpPkg)
		client, err := algod.GetClient(dataDir)
		cobra.CheckErr(err)

		status, response, err := algod.NewStatus(ctx, client, httpPkg)
		utils.WithInvalidResponsesExplanations(err, response, cmd.UsageString())

		if status.State == algod.FastCatchupState {
			log.Fatal(style.Red.Render("Node is currently catching up."))
		}

		// Get the latest catchpoint
		catchpoint, _, err := algod.GetLatestCatchpoint(httpPkg, status.Network)
		if err != nil && err == api.ErrInvalidNetwork {
			log.Fatal("This network does not support fast-catchup.")
		} else {
			log.Info(style.Green.Render("Latest Catchpoint: " + catchpoint))
		}

		// Start catchup
		res, _, err := algod.StartCatchup(ctx, client, catchpoint, nil)
		if err != nil {
			log.Fatal(err)
		}

		log.Info(style.Green.Render(res))
	},
}, &dataDir)
