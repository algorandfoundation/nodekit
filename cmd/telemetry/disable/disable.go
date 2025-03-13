package disable

import (
	"github.com/algorandfoundation/nodekit/cmd/utils/explanations"
	"github.com/algorandfoundation/nodekit/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var Short = "Disable Telemetry"
var Long = lipgloss.JoinVertical(
	lipgloss.Left,
	style.Purple(style.BANNER),
	"",
	style.Bold(Short),
	"",
	style.BoldUnderline("Overview:"),
	"Disable telemetry for the Algorand daemon.",
	"",
	style.Yellow.Render(explanations.ExperimentalWarning),
)
var Cmd = &cobra.Command{
	Use:   "disable",
	Short: Short,
	Long:  Long,
}

func init() {
	Cmd.AddCommand(nodelyCmd)
}
