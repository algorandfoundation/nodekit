package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/algorandfoundation/nodekit/api"
	"github.com/algorandfoundation/nodekit/cmd/utils"
	"github.com/algorandfoundation/nodekit/cmd/utils/explanations"
	"github.com/algorandfoundation/nodekit/internal/algod"
	"github.com/algorandfoundation/nodekit/internal/system"
	"github.com/algorandfoundation/nodekit/ui/style"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// UpgradeMsg is a constant string used to indicate the start of the Algod upgrade process.
const UpgradeMsg = "Upgrading Algod"

const NodeKitUpgradeSuccessMsg = "NodeKit upgraded successfully. Restarting with the latest version."

// AlgorandUpgradeSuccessMsg is shown when the upgrade ran but the installed version could not be read.
const AlgorandUpgradeSuccessMsg = "Algorand upgrade finished."

// AlgorandUpgradedMsg is shown when the package moved to a new version.
const AlgorandUpgradedMsg = "Algorand upgraded from %s to %s."

// AlgorandUpToDateMsg is shown when there was nothing newer to install.
const AlgorandUpToDateMsg = "Algorand is already up to date (%s)."

var upgradeShort = "Upgrade the node daemon"

// acceptNewKey installs a new Algorand repository signing key without asking.
var acceptNewKey bool

var upgradeLong = lipgloss.JoinVertical(
	lipgloss.Left,
	style.Purple(style.BANNER),
	"",
	style.Bold(upgradeShort),
	"",
	style.BoldUnderline("Overview:"),
	"Upgrade Algorand packages if it was installed with package manager.",
	"",
	style.Yellow.Render("This requires the daemon to be installed on your system."),
)

// upgradeCmd is a Cobra command used to upgrade Algod, utilizing the OS-specific package manager if applicable.
var upgradeCmd = &cobra.Command{
	Use:          "upgrade",
	Short:        upgradeShort,
	Long:         upgradeLong,
	SilenceUsage: true,
	Run: func(cmd *cobra.Command, args []string) {
		if NeedsUpgrade && !system.IsReexec() {
			log.Info(style.Green.Render("Upgrading NodeKit"))
			executable, err := os.Executable()
			if err != nil {
				log.Fatal(err)
			}

			err = system.Upgrade(new(api.HttpPkg))
			if err != nil {
				log.Fatal(err)
			}
			log.Info(style.Green.Render(NodeKitUpgradeSuccessMsg))
			if err := system.Reexec(executable, os.Args[1:]); err != nil {
				log.Fatal(err)
			}
			return
		}

		log.Info(style.Green.Render(UpgradeMsg))
		// Warn user for prompt
		log.Warn(style.Yellow.Render(explanations.SudoWarningMsg))
		before, _ := algod.InstalledVersion()
		err := algod.Update(confirmNewKey)
		if err != nil {
			log.Fatal(err)
		}
		after, _ := algod.InstalledVersion()
		switch {
		case before != "" && before == after:
			log.Info(style.Green.Render(fmt.Sprintf(AlgorandUpToDateMsg, after)))
		case before != "" && after != "":
			log.Info(style.Green.Render(fmt.Sprintf(AlgorandUpgradedMsg, before, after)))
			// Give the package's service restart time to settle
			time.Sleep(5 * time.Second)
		default:
			log.Info(style.Green.Render(AlgorandUpgradeSuccessMsg))
			time.Sleep(5 * time.Second)
		}

		// If it's not running, start the daemon (can happen)
		if !algod.IsRunning(algodData) {
			err = algod.Start()
			if err != nil {
				log.Error(err)
				os.Exit(1)
			}
		}
	},
}

// confirmNewKey asks before trusting a new repository signing key, unless
// --accept-new-key was passed. nodekit only offers keys under the Algorand
// primary key it expects.
func confirmNewKey(question string) bool {
	if acceptNewKey {
		log.Info("Accepting the new key (--accept-new-key)")
		return true
	}
	return utils.Confirm(question)
}

func init() {
	upgradeCmd.Flags().BoolVar(&acceptNewKey, "accept-new-key", false, style.Yellow.Render("install a new Algorand package signing key without asking"))
}
