package configure

import (
	"fmt"
	"time"

	cmdutils "github.com/algorandfoundation/nodekit/cmd/utils"
	"github.com/algorandfoundation/nodekit/cmd/utils/explanations"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/internal/algod/config"
	"github.com/algorandfoundation/nodekit/internal/algod/utils"
	"github.com/algorandfoundation/nodekit/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var enableHybrid bool

// algodShort provides a brief description of the algod command, emphasizing its role in installing algod files.
var algodShort = "Configure options for the Algorand daemon."

// algodLong provides a detailed description of the algod command, its purpose, and an experimental warning note.
var algodLong = lipgloss.JoinVertical(
	lipgloss.Left,
	style.Purple(style.BANNER),
	"",
	style.Bold(algodShort),
	"",
	style.BoldUnderline("Overview:"),
	"Modify various configuration options available for the Algorand daemon.",
)

// TODO: Check if we should enforce sudo for this.
// algodCmd is a Cobra command for managing Algorand configuration
var algodCmd = cmdutils.WithAlgodFlags(&cobra.Command{
	Use:   "algod",
	Short: algodShort,
	Long:  algodLong,
	RunE: func(cmd *cobra.Command, args []string) error {
		dataDir, err := algod.GetDataDir(algodData)
		if err != nil {
			log.Fatal(err)
		}

		// Current Node Configuration
		currentConfig, _ := utils.GetConfigFromDataDir(dataDir)

		// OR (`||`) additional flags for `hasFlags` when adding something new.
		hasHybrid := cmd.Flags().Lookup("hybrid").Changed
		hasFlags := hasHybrid

		restartRequired := false

		// Are we doing something? If not, just display the current configuration.
		if hasFlags {
			newConfig := &config.Config{
				// EnableP2PHybridMode: currentConfig.EnableP2PHybridMode,
			}

			if hasHybrid {
				newConfig.EnableP2PHybridMode = &enableHybrid
			}

			mergedConfig := config.MergeAlgodConfigs(*currentConfig, *newConfig)
			if currentConfig.IsEqual(mergedConfig) {
				log.Debug("Configuration up to date, nothing to do")
			} else {
				err := utils.WriteConfigToDataDir(dataDir, &mergedConfig)
				if err != nil {
					log.Warnf("%s", err)
					log.Fatalf("%s", explanations.AlgorandPermissionErrorMsg)
				}
				restartRequired = true
			}

		} else {

			hybridModeStatus := "Disabled"
			if currentConfig.EnableP2PHybridMode != nil && *currentConfig.EnableP2PHybridMode {
				hybridModeStatus = "Enabled"
			}

			rows := [][]string{
				{"EnableP2PHybridMode:", hybridModeStatus},
			}

			var (
				cellStyle      = lipgloss.NewStyle().Padding(0)
				optionRowStyle = cellStyle.Align(lipgloss.Right)
				valueRowStyle  = cellStyle.Align(lipgloss.Left)
			)

			configurationTable := table.New().
				Border(lipgloss.HiddenBorder()).
				StyleFunc(func(row, col int) lipgloss.Style {
					if col == 0 {
						return optionRowStyle
					}
					return valueRowStyle
				}).
				Rows(rows...)

			currentConfiguration := lipgloss.JoinVertical(
				lipgloss.Left,
				style.BoldUnderline("Current Configuration:"),
				configurationTable.String(),
			)

			fmt.Println(currentConfiguration)
		}

		if restartRequired {
			log.Debug("Restarting node...")
			err = algod.Stop()
			if err != nil {
				log.Fatal(err)
			}

			// Wait 1 second.
			// Calling stop & start too quickly on Mac (launchctl) appears to
			// result in a false successfully start. Haven't investigated why.
			time.Sleep(1 * time.Second)

			err = algod.Start()
			if err != nil {
				log.Fatal(err)
			}
			log.Debug("Node restarted successfully.")
		}
		return nil
	},
}, &algodData)

func init() {
	algodCmd.Flags().BoolVar(&enableHybrid, "hybrid", true, "Enable or Disable P2P Hybrid Mode")
}
