package catchup

import (
	"github.com/algorandfoundation/algorun-tui/cmd/utils"
	"github.com/algorandfoundation/algorun-tui/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	// endpoint is a string variable used to store the algod API endpoint address for communication with the node.
	endpoint string = ""

	// token is a string flag used to store the admin token required for authenticating with the Algod API.
	token string = ""

	// force indicates whether to bypass certain checks or enforcement logic within a function or command execution flow.
	force bool = false

	// cmdLong provides a detailed description of the Fast-Catchup feature, explaining its purpose and expected sync durations.
	cmdLong = lipgloss.JoinVertical(
		lipgloss.Left,
		style.Purple(style.BANNER),
		"",
		style.Bold("Fast-Catchup is a feature that allows your node to catch up to the network faster than normal."),
		"",
		style.BoldUnderline("Overview:"),
		"The entire process should sync a node in minutes rather than hours or days.",
		"Actual sync times may vary depending on the number of accounts, number of blocks and the network.",
		"",
		style.Yellow.Render("Note: Not all networks support Fast-Catchup."),
	)

	// Cmd represents the root command for managing an Algorand node, including its description and usage instructions.
	Cmd = utils.WithAlgodFlags(&cobra.Command{
		Use:   "catchup",
		Short: "Manage Fast-Catchup for your node",
		Long:  cmdLong,
	}, &endpoint, &token)
)

func init() {
	Cmd.AddCommand(startCmd)
	Cmd.AddCommand(stopCmd)
	Cmd.AddCommand(debugCmd)
}
