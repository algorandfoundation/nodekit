package configure

import (
	"errors"
	"github.com/algorandfoundation/algorun-tui/cmd/utils/explanations"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/algorandfoundation/algorun-tui/internal/system"
	"github.com/algorandfoundation/algorun-tui/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var serviceShort = "Install service files for the Algorand daemon."
var serviceLong = lipgloss.JoinVertical(
	lipgloss.Left,
	style.Purple(style.BANNER),
	"",
	style.Bold(serviceShort),
	"",
	style.BoldUnderline("Overview:"),
	"Ensuring that the Algorand daemon is installed and running as a service.",
	"",
	style.Yellow.Render(explanations.ExperimentalWarning),
)
var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: serviceShort,
	Long:  serviceLong,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if !system.IsSudo() {
			return errors.New(
				"you need to be root to run this command. Please run this command with sudo")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Combine this with algod.UpdateService and algod.SetNetwork
		return algod.EnsureService()
	},
}
