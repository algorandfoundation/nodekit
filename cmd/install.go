package cmd

import (
	"github.com/algorandfoundation/nodekit/cmd/utils"
	"github.com/algorandfoundation/nodekit/cmd/utils/explanations"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
	"time"
)

// InstallMsg is a constant string used to indicate the start of the Algorand installation process with a specific message.
const InstallMsg = "Installing Algorand"

// InstallExistsMsg is a constant string used to indicate that the Algod is already installed on the system.
const InstallExistsMsg = "algod is already installed"

var installShort = "Install the node daemon"

var installLong = lipgloss.JoinVertical(
	lipgloss.Left,
	style.Purple(style.BANNER),
	"",
	style.Bold(installShort),
	"",
	style.BoldUnderline("Overview:"),
	"Configures the local package manager and installs the algorand daemon on your local machine",
	"",
)

// installCmd is a Cobra command that installs the Algorand daemon on the local machine, ensuring the service is operational.
var installCmd = utils.WithAlgodFlags(&cobra.Command{
	Use:          "install",
	Short:        installShort,
	Long:         installLong,
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info(style.Green.Render(InstallMsg))
		log.Warn(style.Yellow.Render(explanations.SudoWarningMsg))

		// Abort when an installation already exists
		if algod.IsInstalled(dataDir) && !force {
			log.Fatal(InstallExistsMsg)
		}

		// Run the installation
		err := algod.Install(dataDir, "mainnet", force)
		if err != nil {
			log.Fatal(err)
		}

		time.Sleep(5 * time.Second)

		// If it's not running, start the daemon (can happen)
		if !algod.IsRunning(dataDir) {
			err = algod.Start(dataDir)
			if err != nil {
				log.Fatal(err)
			}
		}

		log.Info(style.Green.Render("Algorand installed successfully 🎉"))
	},
}, &dataDir)

func init() {
	installCmd.Flags().BoolVarP(&force, "force", "f", false, style.Yellow.Render("forcefully install the node"))
}
