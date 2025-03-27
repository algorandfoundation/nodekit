package cmd

import (
	"context"
	"fmt"
	"github.com/algorandfoundation/nodekit/api"
	"github.com/algorandfoundation/nodekit/cmd/catchup"
	"github.com/algorandfoundation/nodekit/cmd/configure"
	"github.com/algorandfoundation/nodekit/cmd/telemetry"
	"github.com/algorandfoundation/nodekit/cmd/utils"
	"github.com/algorandfoundation/nodekit/cmd/utils/explanations"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/internal/system"
	"github.com/algorandfoundation/nodekit/ui"
	"github.com/algorandfoundation/nodekit/ui/app"
	"github.com/algorandfoundation/nodekit/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"runtime"
)

var (
	Name = "nodekit"

	NeedsUpgrade = false

	// whether the user forbids incentive eligibility fees to be set
	IncentivesDisabled = false

	// ForceQrCode determines whether QR code generation is enforced, typically for displaying node-related information.
	ForceQrCode = false

	// algodEndpoint defines the URI address of the Algorand node, including the protocol (http/https), for client communication.
	algodData string

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
		Use:   Name,
		Short: short,
		Long:  long,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		Run: func(cmd *cobra.Command, args []string) {
			log.SetOutput(cmd.OutOrStdout())
			err := runTUI(cmd, algodData, IncentivesDisabled, ForceQrCode, cmd.Version)
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
	RootCmd.Flags().BoolVarP(&IncentivesDisabled, "no-incentives", "n", false, style.LightBlue("Disable setting incentive eligibility fees"))
	RootCmd.Flags().BoolVarP(&ForceQrCode, "force-qr", "", false, "Enable QR and bypass guards")
	RootCmd.SetVersionTemplate(fmt.Sprintf("nodekit-%s-%s@{{.Version}}\n", runtime.GOARCH, runtime.GOOS))
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
		RootCmd.AddCommand(telemetry.Cmd)
	}
}

// Execute executes the root command.
func Execute(version string, needsUpgrade bool) error {
	RootCmd.Version = version
	NeedsUpgrade = needsUpgrade
	return RootCmd.Execute()
}

func runTUI(cmd *cobra.Command, dataDir string, incentivesFlag bool, forceQrFlag bool, version string) error {
	if cmd == nil {
		return fmt.Errorf("cmd is nil")
	}
	// Create the dependencies
	ctx := context.Background()
	httpPkg := new(api.HttpPkg)
	t := new(system.Clock)
	client, err := algod.GetClient(dataDir)
	cobra.CheckErr(err)

	// Fetch the state and handle any creation errors
	state, stateResponse, err := algod.NewStateModel(ctx, client, httpPkg, incentivesFlag, forceQrFlag, version)
	utils.WithInvalidResponsesExplanations(err, stateResponse, cmd.UsageString())
	cobra.CheckErr(err)
	// Construct the TUI Model from the State
	m, err := ui.NewViewportViewModel(state)
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
		// Check if the instance is lagging
		lagging, err := algod.IsLagging(httpPkg, state.Status.LastRound, state.Status.Network)
		if err != nil {
			cobra.CheckErr(err)
		}
		state.Watch(func(status *algod.StateModel, err error) {
			// Handle Fast Catchup
			if state.Status.State == algod.FastCatchupState {
				p.Send(app.CatchupModal)
				lagging = false
			} else if lagging && status.Status.State != algod.FastCatchupState {
				p.Send(app.LaggingModal)
				lagging = false
			}

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
	return err
}
