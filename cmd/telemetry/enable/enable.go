package enable

import (
	"github.com/algorandfoundation/nodekit/cmd/utils/explanations"
	"github.com/algorandfoundation/nodekit/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var Short = "Enable Telemetry"
var Long = lipgloss.JoinVertical(
	lipgloss.Left,
	style.Purple(style.BANNER),
	"",
	style.Bold(Short),
	"",
	style.BoldUnderline("Overview:"),
	"Configure telemetry for the Algorand daemon.",
	"",
	style.Yellow.Render(explanations.ExperimentalWarning),
)
var Cmd = &cobra.Command{
	Use:   "enable",
	Short: Short,
	Long:  Long,
}

func init() {
	Cmd.AddCommand(nodelyCmd)
}
