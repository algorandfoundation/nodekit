package telemetry

import (
	"github.com/algorandfoundation/nodekit/cmd/telemetry/disable"
	"github.com/algorandfoundation/nodekit/cmd/telemetry/enable"
	"github.com/algorandfoundation/nodekit/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var Short = "Configure telemetry profile"
var Long = lipgloss.JoinVertical(
	lipgloss.Left,
	style.Purple(style.BANNER),
	"",
	style.Bold(Short),
	"",
	style.BoldUnderline("Overview:"),
	"Enable, disable and view telemetry status.",
	"",
)

var Cmd = &cobra.Command{
	Use:   "telemetry",
	Short: Short,
	Long:  Long,
}

func init() {
	Cmd.AddCommand(enable.Cmd)
	Cmd.AddCommand(disable.Cmd)
	Cmd.AddCommand(statusCmd)
}
