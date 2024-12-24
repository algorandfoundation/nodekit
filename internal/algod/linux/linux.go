package linux

import (
	"bytes"
	"fmt"
	"github.com/algorandfoundation/algorun-tui/internal/algod/fallback"
	"github.com/algorandfoundation/algorun-tui/internal/system"
	"github.com/charmbracelet/log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"text/template"
)

// PackageManagerNotFoundMsg is an error message indicating the absence of a supported package manager for uninstalling Algorand.
const PackageManagerNotFoundMsg = "could not find a package manager to uninstall Algorand"

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
func Install() error {
	log.Info("Installing Algod on Linux")
	// Based off of https://developer.algorand.org/docs/run-a-node/setup/install/#installation-with-a-package-manager
	if system.CmdExists("apt-get") { // On some Debian systems we use apt-get
		log.Info("Installing with apt-get")
		return system.RunAll(append(InstallRequirements(), system.CmdsList{
			{"sudo", "apt-get", "update"},
			{"sudo", "apt-get", "install", "-y", "gnupg2", "curl", "software-properties-common"},
			{"sh", "-c", "curl -o - https://releases.algorand.com/key.pub | sudo tee /etc/apt/trusted.gpg.d/algorand.asc"},
			{"sudo", "add-apt-repository", "-y", fmt.Sprintf("deb [arch=%s] https://releases.algorand.com/deb/ stable main", runtime.GOARCH)},
			{"sudo", "apt-get", "update"},
			{"sudo", "apt-get", "install", "-y", "algorand"},
		}...))
	}

	if system.CmdExists("dnf") { // On Fedora and CentOs8 there's the dnf package manager
		log.Printf("Installing with dnf")
		return system.RunAll(append(InstallRequirements(), system.CmdsList{
			{"curl", "-O", "https://releases.algorand.com/rpm/rpm_algorand.pub"},
			{"sudo", "rpmkeys", "--import", "rpm_algorand.pub"},
			{"sudo", "dnf", "install", "-y", "dnf-command(config-manager)"},
			{"sudo", "dnf", "config-manager", "--add-repo=https://releases.algorand.com/rpm/stable/algorand.repo"},
			{"sudo", "dnf", "install", "-y", "algorand"},
			{"sudo", "systemctl", "enable", "algorand.service"},
			{"sudo", "systemctl", "start", "algorand.service"},
			{"rm", "-f", "rpm_algorand.pub"},
		}...))

	}

	// TODO: watch this method to see if it is ever used
	return fallback.Install()
}

// Uninstall removes the Algorand software using a supported package manager or clears related system files if necessary.
// Returns an error if a supported package manager is not found or if any command fails during execution.
func Uninstall() error {
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
func Start() error {
	return exec.Command("sudo", "systemctl", "start", "algorand").Run()
}

// Stop shuts down the Algorand algod system process on Linux using the systemctl stop command.
// Returns an error if the operation fails.
// TODO: Replace with D-Bus integration
func Stop() error {
	return exec.Command("sudo", "systemctl", "stop", "algorand").Run()
}

// IsService checks if the "algorand.service" is listed as a systemd unit file on Linux.
// Returns true if it exists.
// TODO: Replace with D-Bus integration
func IsService() bool {
	out, err := system.Run([]string{"sudo", "systemctl", "list-unit-files", "algorand.service"})
	if err != nil {
		return false
	}
	return strings.Contains(out, "algorand.service")
}

// UpdateService updates the systemd service file for the Algorand daemon
// with a new data directory path and reloads the daemon.
func UpdateService(dataDirectoryPath string) error {

	algodPath, err := exec.LookPath("algod")
	if err != nil {
		fmt.Printf("Failed to find algod binary: %v\n", err)
		os.Exit(1)
	}

	// Path to the systemd service override file
	// Assuming that this is the same everywhere systemd is used
	overrideFilePath := "/etc/systemd/system/algorand.service.d/override.conf"

	// Create the override directory if it doesn't exist
	err = os.MkdirAll("/etc/systemd/system/algorand.service.d", 0755)
	if err != nil {
		fmt.Printf("Failed to create override directory: %v\n", err)
		os.Exit(1)
	}

	// Content of the override file
	const overrideTemplate = `[Unit]
Description=Algorand daemon {{.AlgodPath}} in {{.DataDirectoryPath}}
[Service]
ExecStart=
ExecStart={{.AlgodPath}} -d {{.DataDirectoryPath}}`

	// Data to fill the template
	data := map[string]string{
		"AlgodPath":         algodPath,
		"DataDirectoryPath": dataDirectoryPath,
	}

	// Parse and execute the template
	tmpl, err := template.New("override").Parse(overrideTemplate)
	if err != nil {
		fmt.Printf("Failed to parse template: %v\n", err)
		os.Exit(1)
	}

	var overrideContent bytes.Buffer
	err = tmpl.Execute(&overrideContent, data)
	if err != nil {
		fmt.Printf("Failed to execute template: %v\n", err)
		os.Exit(1)
	}

	// Write the override content to the file
	err = os.WriteFile(overrideFilePath, overrideContent.Bytes(), 0644)
	if err != nil {
		fmt.Printf("Failed to write override file: %v\n", err)
		os.Exit(1)
	}

	// Reload systemd manager configuration
	cmd := exec.Command("systemctl", "daemon-reload")
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Failed to reload systemd daemon: %v\n", err)
		os.Exit(1)
	}

	log.Info("Algorand service file updated successfully.")

	return nil
}
