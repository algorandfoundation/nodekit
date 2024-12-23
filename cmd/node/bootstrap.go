package node

import (
	"fmt"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/algorandfoundation/algorun-tui/ui/app"
	"github.com/algorandfoundation/algorun-tui/ui/bootstrap"
	"github.com/algorandfoundation/algorun-tui/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"time"
)

var in = `# Welcome!

This is the beginning of your adventure into running the an Algorand node!

Morbi mauris quam, ornare ac commodo et, posuere id sem. Nulla id condimentum mauris. In vehicula sit amet libero vitae interdum. Nullam ac massa in erat volutpat sodales. Integer imperdiet enim cursus, ullamcorper tortor vel, imperdiet diam. Maecenas viverra ex iaculis, vehicula ligula quis, cursus lorem. Mauris nec nunc feugiat tortor sollicitudin porta ac quis turpis. Nam auctor hendrerit metus et pharetra.

`

// bootstrapCmd defines the "debug" command used to display diagnostic information for developers, including debug data.
var bootstrapCmd = &cobra.Command{
	Use:          "bootstrap",
	Short:        "Initialize a fresh node. Alias for install, catchup, and start.",
	Long:         "Text",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print(style.Purple(style.BANNER))
		out, err := glamour.Render(in, "dark")
		if err != nil {
			return err
		}
		fmt.Println(out)

		model := bootstrap.NewModel()
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

		log.Warn(style.Yellow.Render(SudoWarningMsg))
		if msg.Install && !algod.IsInstalled() {
			err := algod.Install()
			if err != nil {
				return err
			}
		}

		// Wait for algod
		time.Sleep(10 * time.Second)

		if !algod.IsRunning() {
			log.Fatal("algod is not running")
		}

		//if msg.Catchup {
		//	ctx := context.Background()
		//	httpPkg := new(api.HttpPkg)
		//	client, err := algod.GetClient(endpoint, token)
		//
		//	// Get the latest catchpoint
		//	catchpoint, _, err := algod.GetLatestCatchpoint(httpPkg, status.Network)
		//	if err != nil && err.Error() == api.InvalidNetworkParamMsg {
		//		log.Fatal("This network does not support fast-catchup.")
		//	} else {
		//		log.Info(style.Green.Render("Latest Catchpoint: " + catchpoint))
		//	}
		//
		//	// Start catchup
		//	res, _, err := algod.StartCatchup(ctx, client, catchpoint, nil)
		//	if err != nil {
		//		log.Fatal(err)
		//	}
		//
		//}
		return nil
	},
}
