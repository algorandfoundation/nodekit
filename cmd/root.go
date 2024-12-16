package cmd

import (
	"context"
	"fmt"
	"github.com/algorandfoundation/algorun-tui/api"
	"github.com/algorandfoundation/algorun-tui/cmd/configure"
	"github.com/algorandfoundation/algorun-tui/cmd/node"
	"github.com/algorandfoundation/algorun-tui/cmd/utils"
	"github.com/algorandfoundation/algorun-tui/internal"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/algorandfoundation/algorun-tui/ui"
	"github.com/algorandfoundation/algorun-tui/ui/explanations"
	"github.com/algorandfoundation/algorun-tui/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"runtime"
	"strings"
)

var (
	algodEndpoint string
	algodToken    = strings.Repeat("a", 64)
	Version       = ""
	rootCmd       = utils.WithAlgodFlags(&cobra.Command{
		Use:   "algorun",
		Short: "Manage Algorand nodes",
		Long:  style.Purple(style.BANNER) + "\n",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			log.SetOutput(cmd.OutOrStdout())
			err := utils.InitConfig()
			if err != nil {
				return err
			}

			if viper.GetString("algod-endpoint") == "" {
				return fmt.Errorf(style.Red.Render("algod-endpoint is required") + explanations.NodeNotFound)
			}

			if viper.GetString("algod-token") == "" {
				return fmt.Errorf(style.Red.Render("algod-token is required"))
			}
			ctx := context.Background()
			client, err := algod.GetClient(viper.GetString("algod-endpoint"), viper.GetString("algod-token"))
			cobra.CheckErr(err)

			httpPkg := new(api.HttpPkg)
			algodStatus, v, err := algod.NewStatus(ctx, client, httpPkg)
			if err != nil {
				return fmt.Errorf(
					style.Red.Render("failed to get status: %s")+explanations.Unreachable,
					err)
			} else if v.StatusCode() == 401 {
				return fmt.Errorf(
					style.Red.Render("failed to get status: Unauthorized") + explanations.TokenInvalid)
			} else if v.StatusCode() != 200 {
				return fmt.Errorf(
					style.Red.Render("failed to get status: error code %d")+explanations.TokenNotAdmin,
					v.StatusCode())
			}

			partkeys, err := internal.GetPartKeys(ctx, client)
			if err != nil {
				return fmt.Errorf(
					style.Red.Render("failed to get participation keys: %s")+
						explanations.TokenNotAdmin,
					err)
			}
			state := internal.StateModel{
				Status: algodStatus,
				Metrics: internal.MetricsModel{
					RoundTime: 0,
					TPS:       0,
					RX:        0,
					TX:        0,
				},
				ParticipationKeys: partkeys,

				Client:  client,
				Context: ctx,
			}
			state.Accounts, err = internal.AccountsFromState(&state, new(internal.Clock), client)
			cobra.CheckErr(err)
			// Fetch current state
			//_, err = state.Status.Fetch(ctx, client, new(internal.HttpPkg))
			//cobra.CheckErr(err)

			m, err := ui.NewViewportViewModel(&state, client)
			cobra.CheckErr(err)

			p := tea.NewProgram(
				m,
				tea.WithAltScreen(),
				tea.WithFPS(120),
			)
			go func() {
				state.Watch(func(status *internal.StateModel, err error) {
					if err == nil {
						p.Send(state)
					}
					if err != nil {
						p.Send(state)
						p.Send(err)
					}
				}, ctx, client)
			}()
			_, err = p.Run()
			return err
		},
	}, &algodEndpoint, &algodToken)
)

func check(err interface{}) {
	if err != nil {
		panic(err)
	}
}

// Handle global flags and set usage templates
func init() {
	log.SetReportTimestamp(false)

	// Configure Version
	if Version == "" {
		Version = "unknown (built from source)"
	}
	rootCmd.Version = Version

	// Add Commands
	rootCmd.AddCommand(statusCmd)
	if runtime.GOOS != "windows" {
		rootCmd.AddCommand(node.Cmd)
		rootCmd.AddCommand(configure.Cmd)
	}
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
