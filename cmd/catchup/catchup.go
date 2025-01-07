package catchup

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

var (
	// dataDir path to the algorand data folder
	dataDir string = ""

	// defaultLag represents the default minimum catchup delay in milliseconds for the Fast Catchup process.
	defaultLag int = 30_000

	// cmdLong provides a detailed description of the Fast-Catchup feature, explaining its purpose and expected sync durations.
	cmdLong = lipgloss.JoinVertical(
		lipgloss.Left,
		style.Purple(style.BANNER),
		"",
		style.Bold("Fast-Catchup is a feature that allows your node to catch up to the network faster than normal."),
		"",
		style.BoldUnderline("Overview:"),
		"The entire process should sync a node in minutes rather than hours or days.",
		"Actual sync times may vary depending on the number of accounts, number of blocks and the network.",
		"",
		style.Yellow.Render("Note: Not all networks support Fast-Catchup."),
	)

	// Cmd represents the root command for managing an Algorand node, including its description and usage instructions.
	Cmd = utils.WithAlgodFlags(&cobra.Command{
		Use:   "catchup",
		Short: "Manage Fast-Catchup for your node",
		Long:  cmdLong,
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
				log.Fatal(style.Red.Render("Node is currently catching up"))
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
)

func init() {
	Cmd.AddCommand(startCmd)
	Cmd.AddCommand(stopCmd)
	Cmd.AddCommand(debugCmd)
}
