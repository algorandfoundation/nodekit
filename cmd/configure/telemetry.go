package configure

import (
	"github.com/algorandfoundation/nodekit/cmd/utils"
	"github.com/algorandfoundation/nodekit/cmd/utils/explanations"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/internal/system"
	"github.com/algorandfoundation/nodekit/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

var dataDir = ""
var telemetryEndpoint string
var telemetryName string

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

var telemetryCmd = utils.WithAlgodFlags(&cobra.Command{
	Use:   "telemetry",
	Short: telemetryShort,
	Long:  telemetryLong,
	Run: func(cmd *cobra.Command, args []string) {
		log.Warn(style.Yellow.Render(explanations.SudoWarningMsg))
		sudoExists := system.CmdExists("sudo")
		if !sudoExists {
			log.Fatal("sudo is not installed, ensure it is available on your system in the PATH.")
		}
		diagExists := system.CmdExists("diagcfg")
		if !diagExists {
			log.Fatal("diagcfg is not installed, ensure it is available on your system in the PATH.")
		}

		dataDir, err := algod.GetDataDir(dataDir)

		err = system.RunAll(system.CmdsList{
			{"sudo", "diagcfg", "-d", dataDir, "telemetry", "endpoint", "-e", telemetryEndpoint},
			{"sudo", "diagcfg", "-d", dataDir, "telemetry", "name", "-n", telemetryName},
			{"sudo", "diagcfg", "-d", dataDir, "telemetry", "enable"},
		})
		if err != nil {
			log.Fatal(err)
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
	telemetryCmd.Flags().StringVarP(&telemetryEndpoint, "endpoint", "e", "https://tel.4160.nodely.io", "Sets the \"URI\" property")
	telemetryCmd.Flags().StringVarP(&telemetryName, "name", "n", "anon", "Enable Algorand remote logging with specified node name")
}
