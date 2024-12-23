package cmd

import (
	"context"
	"github.com/algorandfoundation/algorun-tui/api"
	"github.com/algorandfoundation/algorun-tui/cmd/node"
	"github.com/algorandfoundation/algorun-tui/cmd/utils"
	"github.com/algorandfoundation/algorun-tui/cmd/utils/explanations"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/algorandfoundation/algorun-tui/internal/system"
	"github.com/algorandfoundation/algorun-tui/ui"
	"github.com/algorandfoundation/algorun-tui/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"runtime"
)

var (

	// algodEndpoint defines the URI address of the Algorand node, including the protocol (http/https), for client communication.
	algodData string

	// Version represents the application version string, which is set during build or defaults to "unknown".
	Version = ""

	short = "Manage Algorand nodes from the command line"
	long  = lipgloss.JoinVertical(
		lipgloss.Left,
		style.Purple(style.BANNER),
		"",
		style.Bold(short),
		"",
		style.BoldUnderline("Overview:"),
		"Welcome to Algorun, a TUI for managing Algorand nodes.",
		"A one stop shop for managing Algorand nodes, including node creation, configuration, and management.",
		"",
		style.Yellow.Render(explanations.ExperimentalWarning),
	)
	// rootCmd is the primary command for managing Algorand nodes, providing CLI functionality and TUI for interaction.
	rootCmd = utils.WithAlgodFlags(&cobra.Command{
		Use:   "algorun",
		Short: short,
		Long:  long,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Run: func(cmd *cobra.Command, args []string) {
			log.SetOutput(cmd.OutOrStdout())
			// Create the dependencies
			ctx := context.Background()
			client, err := algod.GetClient("/var/lib/algorand")
			cobra.CheckErr(err)
			httpPkg := new(api.HttpPkg)
			t := new(system.Clock)
			// Fetch the state and handle any creation errors
			state, stateResponse, err := algod.NewStateModel(ctx, client, httpPkg)
			utils.WithInvalidResponsesExplanations(err, stateResponse, cmd.UsageString())
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
				state.Watch(func(status *algod.StateModel, err error) {
					if err == nil {
						p.Send(state)
					}
					if err != nil {
						p.Send(state)
						p.Send(err)
					}
				}, ctx, t)
			}()

			// Execute the TUI Application
			_, err = p.Run()
			if err != nil {
				log.Fatal(err)
			}
		},
	}, &algodData)
)

// init initializes the application, setting up logging, commands, and version information.
func init() {
	log.SetReportTimestamp(false)

	// Configure Version
	if Version == "" {
		Version = "unknown (built from source)"
	}
	rootCmd.Version = Version

	// Add Commands
	if runtime.GOOS != "windows" {
		rootCmd.AddCommand(node.Cmd)
	}
}

// Execute executes the root command.
func Execute() error {
	return rootCmd.Execute()
}
