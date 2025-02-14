package configure

import (
	"github.com/algorandfoundation/nodekit/cmd/utils"
	"github.com/algorandfoundation/nodekit/cmd/utils/explanations"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// serviceShort provides a brief description of the service command, emphasizing its role in installing service files.
var serviceShort = "Install service files for the Algorand daemon."

// serviceLong provides a detailed description of the service command, its purpose, and an experimental warning note.
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

// serviceCmd is a Cobra command for managing Algorand service files, requiring root privileges to ensure proper execution.
var serviceCmd = &cobra.Command{
	Use:               "service",
	Short:             serviceShort,
	Long:              serviceLong,
	PersistentPreRunE: utils.IsSudoCmd,
	RunE: func(cmd *cobra.Command, args []string) error {
		// TODO: Combine this with algod.UpdateService and algod.SetNetwork
		return algod.EnsureService()
	},
}
