package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/algorandfoundation/hack-tui/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

// statusCmd is the main entrypoint for the `status` cobra.Command with a tea.Program
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get the node status",
	Long:  ui.Purple(BANNER) + "\n" + ui.LightBlue("View the node status"),
	RunE: func(cmd *cobra.Command, args []string) error {
		if viper.GetString("server") == "" {
			return errors.New(ui.Magenta("server is required"))
		}

		// Get Algod from configuration
		client, err := getClient()
		cobra.CheckErr(err)

		// Create the TUI
		view, err := ui.MakeStatusViewModel(context.Background(), client)

		cobra.CheckErr(err)
		p := tea.NewProgram(view, tea.WithAltScreen())

		// Execute the Command
		if _, err := p.Run(); err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
		return nil
	},
}
