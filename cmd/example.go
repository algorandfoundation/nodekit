package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/algorandfoundation/hack-tui/internal"
	"github.com/algorandfoundation/hack-tui/ui"
	"github.com/algorandfoundation/hack-tui/ui/pages/keys"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

// exampleCmd is the main entrypoint for the `status` cobra.Command with a tea.Program
var exampleCmd = &cobra.Command{
	Use:   "example",
	Short: "Get the node status",
	Long:  ui.Purple(BANNER) + "\n" + ui.LightBlue("View the node status"),
	RunE: func(cmd *cobra.Command, args []string) error {
		if viper.GetString("server") == "" {
			return errors.New(ui.Magenta("server is required"))
		}

		// Get Algod from configuration
		client, err := getClient()
		cobra.CheckErr(err)

		partkeys, err := internal.GetPartKeys(context.Background(), client)
		cobra.CheckErr(err)

		// Create the TUI
		view := keys.New((*partkeys)[0].Address, partkeys)
		p := tea.NewProgram(view, tea.WithAltScreen())

		// Execute the Command
		if _, err := p.Run(); err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
		return nil
	},
}
