package linux

import (
	"errors"
	"fmt"
	"github.com/algorandfoundation/nodekit/internal/algod/fallback"
	"github.com/algorandfoundation/nodekit/internal/system"
	"github.com/charmbracelet/log"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// PackageManagerNotFoundMsg is an error message indicating the absence of a supported package manager for uninstalling Algorand.
const PackageManagerNotFoundMsg = "could not find a package manager to uninstall Algorand"
const SystemdOverridesDir = "/etc/systemd/system/algorand.service.d"
const SystemdOverrideFilePath = SystemdOverridesDir + "/override.conf"

// Algod represents an implementation of the system.Interface tailored for managing the Algod service.
// It includes details about the service's executable path and associated data directory.
type Algod struct {
	system.Interface
	Path              string
	DataDirectoryPath string
}

// InstallRequirements generates installation commands for "sudo" based on the detected package manager and system state.
func InstallRequirements() system.CmdsList {
	var cmds system.CmdsList
	if (system.CmdExists("sudo") && system.CmdExists("prep")) || os.Geteuid() != 0 {
		return cmds
	}
	if system.CmdExists("apt-get") {
		return system.CmdsList{
			{"apt-get", "update"},
			{"apt-get", "install", "-y", "sudo", "procps"},
		}
	}

	if system.CmdExists("dnf") {
		return system.CmdsList{
			{"dnf", "install", "-y", "sudo", "procps-ng"},
		}
	}
	return cmds
}

// Install installs Algorand development tools or node software depending on the package manager.
func Install(dataDir string, network string, force bool) error {
	log.Info("Installing Algod on Linux")
	// Based off of https://developer.algorand.org/docs/run-a-node/setup/install/#installation-with-a-package-manager
	if system.CmdExists("apt-get") { // On some Debian systems we use apt-get
		log.Info("Installing with apt-get")
		err := system.RunAll(append(InstallRequirements(), system.CmdsList{
			{"sudo", "apt-get", "update"},
			{"sudo", "apt-get", "install", "-y", "gnupg2", "curl", "software-properties-common"},
			{"sh", "-c", "curl -o - https://releases.algorand.com/key.pub | sudo tee /etc/apt/trusted.gpg.d/algorand.asc"},
			{"sudo", "add-apt-repository", "-y", fmt.Sprintf("deb [arch=%s] https://releases.algorand.com/deb/ stable main", runtime.GOARCH)},
			{"sudo", "apt-get", "update"},
			{"sudo", "apt-get", "install", "-y", "algorand"},
		}...))
		if err != nil {
			return err
		}
	}

	if system.CmdExists("dnf") { // On Fedora and CentOs8 there's the dnf package manager
		log.Printf("Installing with dnf")
		err := system.RunAll(append(InstallRequirements(), system.CmdsList{
			{"curl", "-O", "https://releases.algorand.com/rpm/rpm_algorand.pub"},
			{"sudo", "rpmkeys", "--import", "rpm_algorand.pub"},
			{"sudo", "dnf", "install", "-y", "dnf-command(config-manager)"},
			{"sudo", "dnf", "config-manager", "--add-repo=https://releases.algorand.com/rpm/stable/algorand.repo"},
			{"sudo", "dnf", "install", "-y", "algorand"},
			{"sudo", "systemctl", "enable", "algorand.service"},
			{"sudo", "systemctl", "start", "algorand.service"},
			{"rm", "-f", "rpm_algorand.pub"},
		}...))
		if err != nil {
			return err
		}
	}

	return fallback.SetNetwork(dataDir, network, force)
}

// Uninstall removes the Algorand software using a supported package manager or clears related system files if necessary.
// Returns an error if a supported package manager is not found or if any command fails during execution.
func Uninstall(dataDir string, force bool) error {
	if dataDir != "" {
		log.Info("Uninstalling Algorand")
		var unInstallCmds system.CmdsList
		// On Ubuntu and Debian there's the apt package manager
		if system.CmdExists("apt-get") {
			log.Info("Using apt-get package manager")
			unInstallCmds = [][]string{
				{"sudo", "apt-get", "autoremove", "algorand", "-y"},
			}
		}
		// On Fedora and CentOs8 there's the dnf package manager
		if system.CmdExists("dnf") {
			log.Info("Using dnf package manager")
			unInstallCmds = [][]string{
				{"sudo", "dnf", "remove", "algorand", "-y"},
			}
		}
		// Error on unsupported package managers
		if len(unInstallCmds) == 0 {
			return fmt.Errorf(PackageManagerNotFoundMsg)
		}

		// Commands to clear systemd algorand.service and any other files, like the configuration override
		unInstallCmds = append(unInstallCmds, []string{"sudo", "bash", "-c", "rm -rf /etc/systemd/system/algorand*"})
		unInstallCmds = append(unInstallCmds, []string{"sudo", "systemctl", "daemon-reload"})

		return system.RunAll(unInstallCmds)
	}
	// TODO: fallback uninstall
	//return fallback.Uninstall(dataDir)
	return errors.New("fallback uninstall not supported on Linux")
}

// Upgrade updates Algorand and its dev tools using an approved package
// manager if available, otherwise returns an error.
func Upgrade() error {
	if system.CmdExists("apt-get") {
		return system.RunAll(system.CmdsList{
			{"sudo", "apt-get", "update"},
			{"sudo", "apt-get", "install", "--only-upgrade", "-y", "algorand"},
		})
	}
	if system.CmdExists("dnf") {
		return system.RunAll(system.CmdsList{
			{"sudo", "dnf", "update", "-y", "--refresh", "algorand"},
		})
	}
	return fmt.Errorf("the *node upgrade* command is currently only available for installations done with an approved package manager. Please use a different method to upgrade")
}

// Start attempts to start the Algorand service using the system's service manager.
// It executes the appropriate command for systemd on Linux-based systems.
// Returns an error if the command fails.
// TODO: Replace with D-Bus integration
func Start(dataDir string) error {
	// Just start the service when no directory is specified
	//if dataDir == "" {
	return exec.Command("sudo", "systemctl", "start", "algorand").Run()
	//}
	//return fallback.Start(dataDir)
}

// Stop shuts down the Algorand algod system process on Linux using the systemctl stop command.
// Returns an error if the operation fails.
// TODO: Replace with D-Bus integration
func Stop(dataDir string) error {
	// Just stop the service when no directory is specified
	if dataDir == "" {
		return exec.Command("sudo", "systemctl", "stop", "algorand").Run()
	}
	// TODO: Deprecate the forceful stopping
	return fallback.Stop(dataDir)
}

// IsService checks if the "algorand.service" is listed as a systemd unit file on Linux.
// Returns true if it exists.
// TODO: Replace with D-Bus integration
func IsService(dataDir string) bool {
	out, err := system.Run([]string{"sudo", "systemctl", "list-unit-files", "algorand*"})
	if err != nil {
		return false
	}
	return strings.Contains(out, "algorand.service")
}
