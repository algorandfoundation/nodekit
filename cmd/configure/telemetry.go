package configure

import (
	cmdutils "github.com/algorandfoundation/nodekit/cmd/utils"
	"github.com/algorandfoundation/nodekit/cmd/utils/explanations"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/internal/algod/telemetry"
	"github.com/algorandfoundation/nodekit/internal/algod/utils"
	"github.com/algorandfoundation/nodekit/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var dataDir = ""
var telemetryEndpoint string
var telemetryName string
var telemetryDisable bool
var telemetryEnable bool

var telemetryShort = "Configure telemetry for the Algorand daemon"
var NodelyTelemetryWarning = "The default telemetry provider is Nodely."
var telemetryLong = lipgloss.JoinVertical(
	lipgloss.Left,
	style.Purple(style.BANNER),
	"",
	style.Bold(telemetryShort),
	"",
	style.BoldUnderline("Overview:"),
	"When a node is run using the algod command, before the script starts the server,",
	"it configures its telemetry based on the appropriate logging.config file.",
	"When a node’s telemetry is enabled, a telemetry state is added to the node’s logger",
	"reflecting the fields contained within the appropriate config file",
	"",
	style.Yellow.Render(NodelyTelemetryWarning),
)

var telemetryCmd = cmdutils.WithAlgodFlags(&cobra.Command{
	Use:               "telemetry",
	Short:             telemetryShort,
	Long:              telemetryLong,
	PersistentPreRunE: cmdutils.IsSudoCmd,
	Run: func(cmd *cobra.Command, args []string) {
		log.Warn(style.Yellow.Render(explanations.SudoWarningMsg))
		resolvedDir, err := algod.GetDataDir(dataDir)
		if err != nil {
			log.Fatal(err)
		}

		// Current Log Configuration
		logConfig, _ := utils.GetLogConfigFromDataDir(resolvedDir)

		hasDisable := cmd.Flags().Lookup("disable").Changed
		hasEnabled := cmd.Flags().Lookup("enable").Changed
		hasEndpoint := cmd.Flags().Lookup("endpoint").Changed
		hasName := cmd.Flags().Lookup("name").Changed

		hasFlags := hasDisable || hasEndpoint || hasName || hasEnabled

		if hasFlags {
			newConfig := telemetry.Config{
				SendToLog:          logConfig.SendToLog,
				GUID:               logConfig.GUID,
				FilePath:           logConfig.FilePath,
				UserName:           logConfig.UserName,
				Password:           logConfig.Password,
				MinLogLevel:        logConfig.MinLogLevel,
				ReportHistoryLevel: logConfig.ReportHistoryLevel,
			}
			if hasEndpoint {
				newConfig.URI = telemetryEndpoint
			} else {
				newConfig.URI = logConfig.URI
			}
			if hasName {
				newConfig.Name = telemetryName
			} else {
				newConfig.Name = logConfig.Name
			}

			if hasDisable {
				newConfig.Enable = false
			} else if hasEnabled {
				newConfig.Enable = true
			} else {
				newConfig.Enable = logConfig.Enable
			}
			mergeConfig := telemetry.MergeLogConfigs(*logConfig, newConfig)
			if logConfig.IsEqual(mergeConfig) {
				log.Debug("Configuration up to date, nothing to do")
			} else {
				logConfig = &mergeConfig
				err := utils.WriteLogConfigToDataDir(resolvedDir, logConfig)
				if err != nil {
					log.Fatal(err)
				}
			}
		}

		log.Debug("Restarting node...")
		err = algod.Stop()
		if err != nil {
			log.Fatal(err)
		}
		err = algod.Start()
		if err != nil {
			log.Fatal(err)
		}
		log.Debug("Node restarted successfully.")
	},
}, &dataDir)

func init() {
	telemetryCmd.Flags().BoolVarP(&telemetryDisable, "disable", "", false, "Disables telemetry")
	telemetryCmd.Flags().BoolVarP(&telemetryEnable, "enable", "", false, "Enables telemetry")
	telemetryCmd.MarkFlagsOneRequired("disable", "enable")
	telemetryCmd.MarkFlagsMutuallyExclusive("disable", "enable")
	telemetryCmd.Flags().StringVarP(&telemetryEndpoint, "endpoint", "e", string(cmdutils.NodelyTelemetryProvider), "Sets the \"URI\" property")
	telemetryCmd.Flags().StringVarP(&telemetryName, "name", "n", "anon", "Enable Algorand remote logging with specified node name")
}
