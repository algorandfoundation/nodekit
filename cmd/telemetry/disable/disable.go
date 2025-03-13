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

var disableQuestion = `
# Overview

**Do you want to disable telemetry with the provider? (y/n)**
`
var Cmd = &cobra.Command{
	Use:   "disable",
	Short: Short,
	Long:  Long,
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
		if !config.Enable {
			log.Fatal(errors.New(NodelyDisabledErrorMsg))
		}
		cmd.Println(style.BANNER)

		answer := cmdutils.Prompt(disableQuestion)
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
}

func init() {
	Cmd.AddCommand(nodelyCmd)
}
