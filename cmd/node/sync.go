package node

import (
	"context"
	"github.com/algorandfoundation/algorun-tui/api"
	"github.com/algorandfoundation/algorun-tui/cmd/utils"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/algorandfoundation/algorun-tui/ui/style"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var syncEndpoint string
var syncToken string
var defaultLag int = 30_000
var syncCmd = utils.WithAlgodFlags(&cobra.Command{
	Use:              "sync",
	Short:            "Fast Catchup",
	Long:             "Checks if the node is caught up and if not, starts catching up.",
	SilenceUsage:     true,
	PersistentPreRun: NeedsToBeRunning,
	Run: func(cmd *cobra.Command, args []string) {
		err := utils.InitConfig()
		if err != nil {
			log.Fatal(err)
		}
		ep := viper.GetString("algod-endpoint")
		t := viper.GetString("algod-token")
		if ep == "" {
			log.Fatal("algod-endpoint is required")
		}
		if t == "" {
			log.Fatal("algod-token is required")
		}
		// TODO: Perf testing as a dedicated cmd (node perf catchup with exit 0 or 1)
		// NOTE: we do not want to pollute this command with perf testing
		// just allow it to post it's minimum requirements and preform a fast catchup if necessary.
		ctx := context.Background()
		httpPkg := new(api.HttpPkg)
		client, err := algod.GetClient(viper.GetString("algod-endpoint"), viper.GetString("algod-token"))
		cobra.CheckErr(err)

		status, _, err := algod.NewStatus(ctx, client, httpPkg)
		if err != nil {
			log.Fatal(err)
		}
		if status.State == algod.FastCatchupState {
			log.Fatal(style.Red.Render("Node is currently catching up. Use --abort to cancel."))
		}

		catchpoint, _, err := algod.GetLatestCatchpoint(httpPkg, status.Network)
		if err != nil {
			log.Fatal(err)
		}
		log.Info(style.Green.Render("Latest Catchpoint: " + catchpoint))

		res, _, err := algod.PostCatchpoint(ctx, client, catchpoint, &api.StartCatchupParams{Min: &defaultLag})
		if err != nil {
			log.Fatal(err)
		}

		log.Info(style.Green.Render(res))
	},
}, &syncEndpoint, &syncToken)
