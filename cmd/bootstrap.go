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

// bootstrapCmd defines the "debug" command used to display diagnostic information for developers, including debug data.
var bootstrapCmd = &cobra.Command{
	Use:          "bootstrap",
	Short:        bootstrapCmdShort,
	Long:         bootstrapCmdLong,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
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

		// Just launch the TUI if it's already running
		if algod.IsInstalled() && algod.IsService() && algod.IsRunning() {
			dir, err := algod.GetDataDir("")
			if err != nil {
				log.Fatal(err)
			}
			err = runTUI(RootCmd, dir, false)
			if err != nil {
				log.Fatal(err)
			}
			return nil
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

		// Create the Bootstrap TUI
		model := bootstrap.NewModel()
		if algod.IsInstalled() {
			model.BootstrapMsg.Install = false
			model.Question = bootstrap.CatchupQuestion
		}
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

			err := algod.Install()
			if err != nil {
				return err
			}

			// Wait for algod
			time.Sleep(10 * time.Second)

			if !algod.IsRunning() {
				log.Fatal("algod is not running. Something went wrong with installation")
			}
		} else {
			if !algod.IsRunning() {
				log.Info(style.Green.Render("Starting Algod 🚀"))
				log.Warn(style.Yellow.Render(explanations.SudoWarningMsg))
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

		// User answer for catchup question
		if msg.Catchup {
			ctx := context.Background()
			httpPkg := new(api.HttpPkg)

			if err != nil {
				return err
			}
			// Create the client
			client, err := algod.GetClient(dataDir)
			if err != nil {
				return err
			}
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

			// Start catchup
			res, _, err := algod.StartCatchup(ctx, client, catchpoint, nil)
			if err != nil {
				log.Fatal(err)
			}
			log.Info(style.Green.Render(res))

		}

		return runTUI(RootCmd, dataDir, false)
	},
}
