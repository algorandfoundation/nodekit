package cmd

import (
	"context"
	"github.com/algorandfoundation/nodekit/api"
	"github.com/algorandfoundation/nodekit/cmd/catchup"
	"github.com/algorandfoundation/nodekit/cmd/configure"
	"github.com/algorandfoundation/nodekit/cmd/utils"
	"github.com/algorandfoundation/nodekit/cmd/utils/explanations"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/internal/system"
	"github.com/algorandfoundation/nodekit/ui"
	"github.com/algorandfoundation/nodekit/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"runtime"
)

var (
	Name = "nodekit"

	// algodEndpoint defines the URI address of the Algorand node, including the protocol (http/https), for client communication.
	algodData string

	// Version represents the application version string, which is set during build or defaults to "unknown".
	Version = ""

	// force indicates whether actions should be performed forcefully, bypassing checks or confirmations.
	force bool = false

	short = "Manage Algorand nodes from the command line"
	long  = lipgloss.JoinVertical(
		lipgloss.Left,
		style.Purple(style.BANNER),
		"",
		style.Bold(short),
		"",
		style.BoldUnderline("Overview:"),
		"Welcome to NodeKit, a TUI for managing Algorand nodes.",
		"A one stop shop for managing Algorand nodes, including node creation, configuration, and management.",
		"",
		style.Yellow.Render(explanations.ExperimentalWarning),
	)
	// RootCmd is the primary command for managing Algorand nodes, providing CLI functionality and TUI for interaction.
	RootCmd = utils.WithAlgodFlags(&cobra.Command{
		Use:     Name,
		Version: Version,
		Short:   short,
		Long:    long,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Run: func(cmd *cobra.Command, args []string) {
			log.SetOutput(cmd.OutOrStdout())
			// Create the dependencies
			ctx := context.Background()
			client, err := algod.GetClient(algodData)
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

// NeedsToBeRunning ensures the Algod software is installed and running before executing the associated Cobra command.
func NeedsToBeRunning(cmd *cobra.Command, args []string) {
	if force {
		return
	}
	if !algod.IsInstalled() {
		log.Fatal(explanations.NotInstalledErrorMsg)
	}
	if !algod.IsRunning() {
		log.Fatal(explanations.NotRunningErrorMsg)
	}
}

// NeedsToBeStopped ensures the operation halts if Algod is not installed or is currently running, unless forced.
func NeedsToBeStopped(cmd *cobra.Command, args []string) {
	if force {
		return
	}
	if !algod.IsInstalled() {
		log.Fatal(explanations.NotInstalledErrorMsg)
	}
	if algod.IsRunning() {
		log.Fatal(explanations.RunningErrorMsg)
	}
}

// init initializes the application, setting up logging, commands, and version information.
func init() {
	log.SetReportTimestamp(false)
	// Add Commands
	if runtime.GOOS != "windows" {
		RootCmd.AddCommand(bootstrapCmd)
		RootCmd.AddCommand(debugCmd)
		RootCmd.AddCommand(installCmd)
		RootCmd.AddCommand(startCmd)
		RootCmd.AddCommand(stopCmd)
		RootCmd.AddCommand(uninstallCmd)
		RootCmd.AddCommand(upgradeCmd)
		RootCmd.AddCommand(catchup.Cmd)
		RootCmd.AddCommand(configure.Cmd)
	}
}

// Execute executes the root command.
func Execute() error {
	return RootCmd.Execute()
}
