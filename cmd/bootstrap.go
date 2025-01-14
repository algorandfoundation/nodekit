package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/algorandfoundation/nodekit/api"
	cmdutils "github.com/algorandfoundation/nodekit/cmd/utils"
	"github.com/algorandfoundation/nodekit/cmd/utils/explanations"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/internal/algod/utils"
	"github.com/algorandfoundation/nodekit/internal/system"
	"github.com/algorandfoundation/nodekit/ui"
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
		ctx := context.Background()
		httpPkg := new(api.HttpPkg)
		r, _ := glamour.NewTermRenderer(
			glamour.WithAutoStyle(),
		)
		fmt.Print(style.Purple(style.BANNER))
		out, err := r.Render(tutorial)
		if err != nil {
			return err
		}
		fmt.Println(out)

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

		if msg == nil {
			return nil
		}

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

		dataDir, err := algod.GetDataDir("")
		if err != nil {
			return err
		}
		// Create the client
		client, err := algod.GetClient(dataDir)
		if err != nil {
			return err
		}

		if msg.Catchup {
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

		t := new(system.Clock)
		// Fetch the state and handle any creation errors
		state, stateResponse, err := algod.NewStateModel(ctx, client, httpPkg)
		cmdutils.WithInvalidResponsesExplanations(err, stateResponse, cmd.UsageString())
		cobra.CheckErr(err)

		// Construct the TUI Model from the State
		m, err := ui.NewViewportViewModel(state, client)
		cobra.CheckErr(err)

		// Construct the TUI Application
		p = tea.NewProgram(
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
		return nil
	},
}
