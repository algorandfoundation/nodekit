package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/algorandfoundation/nodekit/api"
	"github.com/algorandfoundation/nodekit/cmd/utils/explanations"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/internal/algod/utils"
	"github.com/algorandfoundation/nodekit/ui/app"
	"github.com/algorandfoundation/nodekit/ui/bootstrap"
	"github.com/algorandfoundation/nodekit/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

const CheckAlgodInterval = 10 * time.Second
const CheckAlgodTimeout = 2 * time.Minute

var CatchpointLagThreshold int = 30_000

// bootstrapCmdShort provides a brief description of the "bootstrap" command to initialize a fresh Algorand node.
var bootstrapCmdShort = "Initialize a fresh node"

// bootstrapCmdLong provides a detailed description of the "bootstrap" command, including its purpose and functionality.
var bootstrapCmdLong = lipgloss.JoinVertical(
	lipgloss.Left,
	style.BANNER,
	"",
	style.Bold(bootstrapCmdShort),
	"",
	style.BoldUnderline("Overview:"),
	"Get up and running with a fresh Algorand node.",
	"Uses the local package manager to install Algorand, and then starts the node and preforms a Fast-Catchup.",
	"",
	style.Yellow.Render("Note: This command only supports the default data directory, /var/lib/algorand"),
)

var tutorial = `# Welcome!

This is the beginning of your adventure into running an Algorand node!

`

var FailedToAutoStartMessage = "Failed to start Algorand automatically."

// bootstrapCmd defines the "debug" command used to display diagnostic information for developers, including debug data.
var bootstrapCmd = &cobra.Command{
	Use:          "bootstrap",
	Short:        bootstrapCmdShort,
	Long:         bootstrapCmdLong,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var client *api.ClientWithResponses
		// Create the Bootstrap TUI
		model := bootstrap.NewModel()
		log.Warn(style.Yellow.Render(explanations.SudoWarningMsg))
		// Try to launch the TUI if it's already running and configured
		if algod.IsInitialized() {
			// Parse the data directory
			dir, err := algod.GetDataDir("")
			if err != nil {
				log.Fatal(err)
			}

			// Wait for the client to respond
			log.Warn(style.Yellow.Render("Waiting for the node to start..."))
			client, err = algod.WaitForClient(context.Background(), dir, CheckAlgodInterval, CheckAlgodTimeout)
			if err != nil {
				log.Fatal(err)
			}

			// Fetch the latest status
			var resp *api.GetStatusResponse
			resp, err = client.GetStatusWithResponse(context.Background())
			// This should not happen, we waited for a status already
			if err != nil {
				log.Fatal(err)
			}
			if resp.StatusCode() != 200 {
				log.Fatal(fmt.Sprintf("Failed to connect to the node at %s", dir))
			}

			// Execute the TUI if we are caught up.
			// TODO: check the delta to see if it is necessary,
			if resp.JSON200.CatchupTime == 0 {
				err = runTUI(RootCmd, dir, false)
				if err != nil {
					log.Fatal(err)
				}
				return nil
			}
		}

		// Exit the application in an invalid state
		if algod.IsInstalled() && !algod.IsService() {
			dataDir, _ := algod.GetDataDir("")
			if dataDir == "" {
				dataDir = "<Path to data directory>"
			}
			log.Warn("algorand is installed, but not running as a service. Continue at your own risk!")
			log.Warn(fmt.Sprintf("try connecting to the node with: ./nodekit -d %s", dataDir))
			log.Fatal("invalid state, exiting")
		}

		// Render the welcome text
		r, _ := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
		)
		fmt.Print(style.Purple(style.BANNER))
		out, err := r.Render(tutorial)
		if err != nil {
			return err
		}
		fmt.Println(out)

		// Ensure it the service is started,
		// in this case we won't be able to query state without the node running
		if algod.IsInstalled() && algod.IsService() && !algod.IsRunning() {
			log.Debug("Algorand is installed, but not running. Attempting to start it automatically.")
			log.Warn(style.Yellow.Render(explanations.SudoWarningMsg))
			err := algod.Start()
			if err != nil {
				log.Error(FailedToAutoStartMessage)
				log.Fatal(err)
			}

		}

		// Prefill questions
		if algod.IsInstalled() {
			model.BootstrapMsg.Install = false
			model.Question = bootstrap.CatchupQuestion
		}
		// Run the Bootstrap TUI
		p := tea.NewProgram(model)
		var msg *app.BootstrapMsg
		go func() {
			for {
				val := <-model.Outside
				switch val.(type) {
				case app.BootstrapMsg:
					msgVal := val.(app.BootstrapMsg)
					msg = &msgVal
				}
			}
		}()
		if _, err := p.Run(); err != nil {
			log.Fatal(err)
		}

		// If the pointer is empty, return (should not happen)
		if msg == nil {
			return nil
		}

		// User Answer for Install Question
		if msg.Install {
			log.Warn(style.Yellow.Render(explanations.SudoWarningMsg))

			// Run the installer
			err := algod.Install()
			if err != nil {
				return err
			}

			// Parse the data directory
			dir, err := algod.GetDataDir("")
			if err != nil {
				log.Fatal(err)
			}

			// Wait for the client to respond
			client, err = algod.WaitForClient(context.Background(), dir, CheckAlgodInterval, CheckAlgodTimeout)
			if err != nil {
				log.Fatal(err)
			}

			if !algod.IsRunning() {
				log.Fatal("algod is not running. Something went wrong with installation")
			}
		} else {
			// This should not happen but just in case, ensure it is running
			if !algod.IsRunning() {
				log.Info(style.Green.Render("Starting Algod 🚀"))
				err := algod.Start()
				if err != nil {
					log.Fatal(err)
				}
				log.Info(style.Green.Render("Algorand started successfully 🎉"))
				time.Sleep(2 * time.Second)
			}
		}

		// Find the data directory automatically
		dataDir, err := algod.GetDataDir("")
		// Wait for the client to respond
		client, err = algod.WaitForClient(context.Background(), dataDir, CheckAlgodInterval, CheckAlgodTimeout)
		if err != nil {
			log.Fatal(err)
		}

		// User answer for catchup question
		if msg.Catchup {
			ctx := context.Background()
			httpPkg := new(api.HttpPkg)
			network, err := utils.GetNetworkFromDataDir(dataDir)
			if err != nil {
				return err
			}
			// Get the latest catchpoint
			catchpoint, _, err := algod.GetLatestCatchpoint(httpPkg, network)
			if err != nil && err.Error() == api.InvalidNetworkParamMsg {
				log.Fatal("This network does not support fast-catchup.")
			} else {
				log.Info(style.Green.Render("Latest Catchpoint: " + catchpoint))
			}

			// Start catchup with round threshold
			res, _, err := algod.StartCatchup(ctx, client, catchpoint, &api.StartCatchupParams{Min: &CatchpointLagThreshold})
			if err != nil {
				log.Fatal(err)
			}
			log.Info(style.Green.Render(res))

		}

		return runTUI(RootCmd, dataDir, false)
	},
}
