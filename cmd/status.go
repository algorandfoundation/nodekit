package cmd

import (
	"fmt"
	"github.com/algorandfoundation/hack-tui/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"os"
)

// statusCmd is the main entrypoint for the `status` cobra.Command with a tea.Program
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Get the node status",
	Long:  ui.Purple(BANNER) + "\n" + ui.LightBlue("View the node status"),
	Run: func(cmd *cobra.Command, args []string) {
		// Get Algod from configuration
		algodClient := getAlgodClient()

		// Create the TUI
		view, err := ui.MakeStatusView(algodClient)
		cobra.CheckErr(err)
		p := tea.NewProgram(view)

		// Execute the Command
		if _, err := p.Run(); err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
	},
}
