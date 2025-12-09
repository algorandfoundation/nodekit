package cmd

import (
	"os"
	"time"

	"github.com/algorandfoundation/nodekit/cmd/utils/explanations"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
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
var installCmd = &cobra.Command{
	Use:          "install",
	Short:        installShort,
	Long:         installLong,
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO: yes flag

		// TODO: get expected version
		log.Info(style.Green.Render(InstallMsg))
		// Warn user for prompt
		log.Warn(style.Yellow.Render(explanations.SudoWarningMsg))

		// TODO: compare expected version to existing version
		if algod.IsInstalled() && !force {
			log.Error(InstallExistsMsg)
			os.Exit(1)
		}

		// Run the installation
		err := algod.Install()
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}

		time.Sleep(5 * time.Second)

		// If it's not running, start the daemon (can happen)
		if !algod.IsRunning(algodData) {
			err = algod.Start()
			if err != nil {
				log.Error(err)
				os.Exit(1)
			}
		}

		log.Info(style.Green.Render("Algorand installed successfully 🎉"))
	},
}

func init() {
	installCmd.Flags().BoolVarP(&force, "force", "f", false, style.Yellow.Render("forcefully install the node"))
}
