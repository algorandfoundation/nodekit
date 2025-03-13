package telemetry

import (
	"fmt"
	cmdutils "github.com/algorandfoundation/nodekit/cmd/utils"
	"github.com/algorandfoundation/nodekit/cmd/utils/explanations"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/internal/algod/utils"
	"github.com/algorandfoundation/nodekit/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var statusShort = "Telemetry status"

var statusLong = lipgloss.JoinVertical(
	lipgloss.Left,
	style.Purple(style.BANNER),
	"",
	style.Bold(statusShort),
	"",
	style.BoldUnderline("Overview:"),
	"Display telemetry profile status for the Algorand daemon.",
	"",
	style.Yellow.Render(explanations.ExperimentalWarning),
)

var dataDir = ""

var statusCmd = cmdutils.WithAlgodFlags(&cobra.Command{
	Use:   "status",
	Short: statusShort,
	Long:  statusLong,
	Run: func(cmd *cobra.Command, args []string) {
		// Resolve Data Directory
		dataDir, err := algod.GetDataDir(dataDir)
		cobra.CheckErr(err)

		// Resolve Configuration
		config, err := utils.GetLogConfigFromDataDir(dataDir)
		cobra.CheckErr(err)

		var action string
		if config.Enable {
			action = "Enabled"
		} else {
			action = "Disabled"
		}

		// Collection to render, starting with the title
		msgs := []string{fmt.Sprintf("Telemetry is %s", style.Bold(action))}

		// Add Provider
		if config.Enable && config.URI == string(cmdutils.NodelyTelemetryProvider) {
			msgs = append(msgs, fmt.Sprintf("Provider: %s", "Nodely"))
		}

		if config.Enable {
			msgs = append(msgs, []string{
				fmt.Sprintf("Node name: %s", config.Name),
				fmt.Sprintf("Telemetry GUID: %s", config.GUID),
				fmt.Sprintf("Telemetry Endpoint: %s", config.URI),
			}...)

		} else {
			msgs = append(msgs, fmt.Sprintf("You can enable nodely telemetry with %s", style.Bold("nodekit telemetry enable nodely")))
		}

		cmd.Println(lipgloss.JoinVertical(
			lipgloss.Left,
			msgs...,
		))
	},
}, &dataDir)
