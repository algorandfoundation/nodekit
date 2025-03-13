package disable

import (
	"errors"
	cmdutils "github.com/algorandfoundation/nodekit/cmd/utils"
	"github.com/algorandfoundation/nodekit/cmd/utils/explanations"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/internal/algod/utils"
	"github.com/algorandfoundation/nodekit/internal/system"
	"github.com/algorandfoundation/nodekit/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"os"
)

var nodelyShort = "Disable Nodely"
var nodelyLong = lipgloss.JoinVertical(
	lipgloss.Left,
	style.Purple(style.BANNER),
	"",
	style.Bold(nodelyShort),
	"",
	style.BoldUnderline("Overview:"),
	"Nodely Telemetry is a free telemetry service offered by a third party (Nodely)",
	"Enabling telemetry will configure your node to send health metrics to Nodely",
	"Privacy note: Information about your node (including participating accounts and approximate geographic location) will be associated with an anonymous user identifier (GUID.)",
	"Tip: Keep this GUID identifier private if you do not want this information to be linked to your identity.",
	"",
	style.Yellow.Render(explanations.ExperimentalWarning),
)

var dataDir string

const (
	NodelyNotConfiguredErrorMsg = "nodely is not configured"
	NodelyDisabledErrorMsg      = "nodely is already disabled"
)

var nodelyQuestion = `
# Overview

Nodely Telemetry is a free telemetry service offered by a third party (Nodely)
Enabling telemetry will configure your node to send health metrics to Nodely

> Privacy note: Information about your node (including participating accounts and approximate geographic location) will be associated with an anonymous user identifier (GUID.)

> Tip: Keep this GUID identifier private if you do not want this information to be linked to your identity.

**Do you want to disable telemetry with the nodely provider? (y/n)**
`

var nodelyCmd = cmdutils.WithAlgodFlags(&cobra.Command{
	Use:   "nodely",
	Short: nodelyShort,
	Long:  nodelyLong,
	Run: func(cmd *cobra.Command, args []string) {
		log.Warn(explanations.SudoWarningMsg)

		// Resolve Data Directory
		dataDir, err := algod.GetDataDir(dataDir)
		if err != nil {
			log.Fatal(err)
		}

		// Fetch configuration
		config, err := utils.GetLogConfigFromDataDir(dataDir)
		cobra.CheckErr(err)

		// Error if already disabled
		if !config.Enable && config.URI == string(cmdutils.NodelyTelemetryProvider) {
			log.Fatal(errors.New(NodelyDisabledErrorMsg))
		}
		cmd.Println(style.BANNER)

		answer := cmdutils.Prompt(nodelyQuestion)
		if answer {
			// Get the path to nodekit
			path, err := os.Executable()
			if err != nil {
				log.Fatal(err)
			}

			// Elevated Configuration
			err = system.RunAll(system.CmdsList{{"sudo", path, "configure", "telemetry", "--datadir", dataDir, "--disable"}})
			if err != nil {
				log.Fatal(err)
			}
		}
	},
}, &dataDir)
