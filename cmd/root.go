package cmd

import (
	"context"
	"fmt"
	"github.com/algorandfoundation/algorun-tui/api"
	"github.com/algorandfoundation/algorun-tui/cmd/configure"
	"github.com/algorandfoundation/algorun-tui/cmd/node"
	"github.com/algorandfoundation/algorun-tui/cmd/utils"
	"github.com/algorandfoundation/algorun-tui/cmd/utils/explanations"
	"github.com/algorandfoundation/algorun-tui/internal"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/algorandfoundation/algorun-tui/ui"
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

			endpoint := viper.GetString("algod-endpoint")
			token := viper.GetString("algod-token")

			if endpoint == "" {
				return fmt.Errorf(style.Red.Render("algod-endpoint is required") + explanations.NodeNotFound)
			}

			if token == "" {
				return fmt.Errorf(style.Red.Render("algod-token is required"))
			}

			// Create the dependencies
			ctx := context.Background()
			client, err := algod.GetClient(endpoint, token)
			cobra.CheckErr(err)
			httpPkg := new(api.HttpPkg)

			// Fetch the state and handle any creation errors
			state, stateResponse, err := internal.NewStateModel(ctx, client, httpPkg)
			if stateResponse.StatusCode() == 401 {
				return fmt.Errorf(
					style.Red.Render("failed to get status: Unauthorized") + explanations.TokenInvalid)
			}
			if stateResponse.StatusCode() > 300 {
				return fmt.Errorf(
					style.Red.Render("failed to get status: error code %d")+explanations.TokenNotAdmin,
					stateResponse.StatusCode())
			}
			cobra.CheckErr(err)

			// Construct the TUI Model from the State
			m, err := ui.NewViewportViewModel(state, client)
			cobra.CheckErr(err)

			// Construct the TUI Application
			p := tea.NewProgram(
				m,
				tea.WithAltScreen(),
				tea.WithFPS(120),
			)

			// Watch for State Updates on a separate thread
			// TODO: refactor into context aware watcher without callbacks
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

			// Execute the TUI Application
			_, err = p.Run()
			return err
		},
	}, &algodEndpoint, &algodToken)
)

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
