package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/algorandfoundation/hack-tui/ui/style"
	"github.com/spf13/cobra"
)

var nodeCmd = &cobra.Command{
	Use:   "node",
	Short: "Algod installation",
	Long:  style.Purple(BANNER) + "\n" + style.LightBlue("View the node status"),
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Algod",
	Long:  "Install Algod on your system",
	Run: func(cmd *cobra.Command, args []string) {
		installNode()
	},
}

var configureCmd = &cobra.Command{
	Use:   "configure",
	Short: "Configure Algod",
	Long:  "Configure Algod settings",
	Run: func(cmd *cobra.Command, args []string) {
		configureNode()
	},
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Algod",
	Long:  "Start Algod on your system (the one on your PATH).",
	Run: func(cmd *cobra.Command, args []string) {
		startNode()
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop Algod",
	Long:  "Stop the Algod process on your system.",
	Run: func(cmd *cobra.Command, args []string) {
		stopNode()
	},
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade Algod",
	Long:  "Upgrade Algod (if installed with package manager).",
	Run: func(cmd *cobra.Command, args []string) {
		upgradeAlgod()
	},
}

func init() {
	rootCmd.AddCommand(nodeCmd)
	nodeCmd.AddCommand(installCmd)
	nodeCmd.AddCommand(configureCmd)
	nodeCmd.AddCommand(startCmd)
	nodeCmd.AddCommand(stopCmd)
	nodeCmd.AddCommand(upgradeCmd)
}

func installNode() {
	fmt.Println("Checking if Algod is installed...")

	// Check if Algod is installed
	if !isAlgodInstalled() {
		fmt.Println("Algod is not installed. Installing...")

		// Install Algod based on OS
		switch runtime.GOOS {
		case "linux":
			installNodeLinux()
		case "darwin":
			installNodeMac()
		default:
			panic("Unsupported OS: " + runtime.GOOS)
		}
	} else {
		fmt.Println("Algod is already installed.")
		printAlgodInfo()
	}

}

func installNodeLinux() {
	fmt.Println("Installing Algod on Linux")

	// Check that we are calling with sudo
	if !isRunningWithSudo() {
		fmt.Println("This command must be run with super-user priviledges (sudo).")
		os.Exit(1)
	}

	var installCmds [][]string
	var postInstallHint string

	// Based off of https://developer.algorand.org/docs/run-a-node/setup/install/#installation-with-a-package-manager

	if checkCmdToolExists("apt") { // On Ubuntu and Debian we use the apt package manager
		fmt.Println("Using apt package manager")
		installCmds = [][]string{
			{"apt", "update"},
			{"apt", "install", "-y", "gnupg2", "curl", "software-properties-common"},
			{"sh", "-c", "curl -o - https://releases.algorand.com/key.pub | tee /etc/apt/trusted.gpg.d/algorand.asc"},
			{"sh", "-c", `add-apt-repository -y "deb [arch=amd64] https://releases.algorand.com/deb/ stable main"`},
			{"apt", "update"},
			{"apt", "install", "-y", "algorand-devtools"},
		}
	} else if checkCmdToolExists("apt-get") { // On some Debian systems we use apt-get
		fmt.Println("Using apt-get package manager")
		installCmds = [][]string{
			{"apt-get", "update"},
			{"apt-get", "install", "-y", "gnupg2", "curl", "software-properties-common"},
			{"sh", "-c", "curl -o - https://releases.algorand.com/key.pub | tee /etc/apt/trusted.gpg.d/algorand.asc"},
			{"sh", "-c", `add-apt-repository -y "deb [arch=amd64] https://releases.algorand.com/deb/ stable main"`},
			{"apt-get", "update"},
			{"apt-get", "install", "-y", "algorand-devtools"},
		}
	} else if checkCmdToolExists("dnf") { // On Fedora and CentOs8 there's the dnf package manager
		fmt.Println("Using dnf package manager")
		installCmds = [][]string{
			{"curl", "-O", "https://releases.algorand.com/rpm/rpm_algorand.pub"},
			{"rpmkeys", "--import", "rpm_algorand.pub"},
			{"dnf", "install", "-y", "dnf-command(config-manager)"},
			{"dnf", "config-manager", "--add-repo=https://releases.algorand.com/rpm/stable/algorand.repo"},
			{"dnf", "install", "-y", "algorand-devtools"},
			{"systemctl", "start", "algorand"},
		}
	} else if checkCmdToolExists("yum") { // On CentOs7 we use the yum package manager
		fmt.Println("Using yum package manager")
		installCmds = [][]string{
			{"curl", "-O", "https://releases.algorand.com/rpm/rpm_algorand.pub"},
			{"rpmkeys", "--import", "rpm_algorand.pub"},
			{"yum", "install", "yum-utils"},
			{"yum-config-manager", "--add-repo", "https://releases.algorand.com/rpm/stable/algorand.repo"},
			{"yum", "install", "-y", "algorand-devtools"},
			{"systemctl", "start", "algorand"},
		}
	} else {
		fmt.Println("Unsupported package manager, possibly due to non-Debian or non-Red Hat based Linux distribution. Will attempt to install using updater script.")
		installCmds = [][]string{
			{"mkdir", "~/node"},
			{"sh", "-c", "cd ~/node"},
			{"wget", "https://raw.githubusercontent.com/algorand/go-algorand/rel/stable/cmd/updater/update.sh"},
			{"chmod", "744", "update.sh"},
			{"sh", "-c", "./update.sh -i -c stable -p ~/node -d ~/node/data -n"},
		}

		postInstallHint = `You may need to add the Algorand binaries to your PATH:
					export ALGORAND_DATA="$HOME/node/data"
					export PATH="$HOME/node:$PATH"
			`
	}

	// Run each installation command
	for _, cmdArgs := range installCmds {
		fmt.Println("Running command:", strings.Join(cmdArgs, " "))
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Command failed: %s\nOutput: %s\nError: %v\n", strings.Join(cmdArgs, " "), output, err)
			cobra.CheckErr(err)
		}
	}

	if postInstallHint != "" {
		fmt.Println(postInstallHint)
	}
}

func installNodeMac() {
	fmt.Println("Installing Algod on macOS...")

	// Based off of the macOS installation instructions
	// https://developer.algorand.org/docs/run-a-node/setup/install/#installing-on-mac

	installCmd := `mkdir ~/node
		cd ~/node
		wget https://raw.githubusercontent.com/algorand/go-algorand/rel/stable/cmd/updater/update.sh
		chmod 744 update.sh
		./update.sh -i -c stable -p ~/node -d ~/node/data -n`

	postInstallHint := `You may need to add the Algorand binaries to your PATH:
		export ALGORAND_DATA="$HOME/node/data"
		export PATH="$HOME/node:$PATH"
	`

	// Run the installation command
	err := exec.Command(installCmd).Run()
	cobra.CheckErr(err)

	if postInstallHint != "" {
		fmt.Println(postInstallHint)
	}
}

// TODO: configure not just data directory but algod path
func configureNode() {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		panic("Unsupported OS: " + runtime.GOOS)
	}

	var systemctlConfigure bool

	// Check systemctl first
	if checkSystemctlAlgorandServiceCreated() {
		if !isRunningWithSudo() {
			fmt.Println("This command must be run with super-user priviledges (sudo).")
			os.Exit(1)
		}

		if promptWrapperYes("Algorand is installed as a service. Do you wish to edit the service file to change the data directory? (y/n)") {
			if checkSystemctlAlgorandServiceActive() {
				fmt.Println("Algorand service is currently running. Please stop the service with *node stop* before editing the service file.")
				os.Exit(1)
			}
			// Edit the service file with the user's new data directory
			systemctlConfigure = true
		} else {
			fmt.Println("Exiting...")
			os.Exit(0)
		}
	}

	// At the end, instead of affectALGORAND_DATA, we'll edit the systemctl algorand.service file
	// i.e., overwrite /etc/systemd/system/algorand.service.d/override.conf
	// ExecStart and Description will be changed to reflect the new data directory
	//

	if !systemctlConfigure {
		fmt.Println("Configuring Data directory for algod started through Algorun...")
	}

	algorandData := os.Getenv("ALGORAND_DATA")

	// Check if ALGORAND_DATA environment variable is set
	if algorandData != "" {
		fmt.Println("ALGORAND_DATA environment variable is set to: " + algorandData)
		fmt.Println("Inspecting the set data directory...")

		if validateAlgorandDataDir(algorandData) {
			fmt.Println("Found valid Algorand Data Directory: " + algorandData)

			if systemctlConfigure {
				if promptWrapperYes("Would you like to set the ALGORAND_DATA env variable as the data directory for the systemd Algorand service? (y/n)") {
					editAlgorandServiceFile(algorandData)
					os.Exit(0)
				}
			}

			if promptWrapperNo("Do you want to set a completely new data directory? (y/n)") {
				fmt.Println("User chose not to set a completely new data directory.")
				os.Exit(0)
			}

			if promptWrapperYes("Do you want to manually input the new data directory? (y/n)") {
				newPath := promptWrapperInput("Enter the new data directory path")

				if !validateAlgorandDataDir(newPath) {
					fmt.Println("Path at ALGORAND_DATA: " + newPath + " is not recognizable as an Algorand Data directory.")
					os.Exit(1)
				}

				if systemctlConfigure {
					// Edit the service file
					editAlgorandServiceFile(newPath)
				} else {
					// Affect the ALGORAND_DATA environment variable
					affectALGORAND_DATA(newPath)
				}
				os.Exit(0)
			}
		} else {
			fmt.Println("Path at ALGORAND_DATA: " + algorandData + " is not recognizable as an Algorand Data directory.")
		}
	} else {
		fmt.Println("ALGORAND_DATA environment variable not set.")
	}

	// Do quick "lazy" check for existing Algorand Data directories
	paths := lazyCheckAlgorandDataDirs()

	if len(paths) != 0 {

		fmt.Println("Quick check found the following potential data directories:")
		for _, path := range paths {
			fmt.Println("✔ " + path)
		}

		if len(paths) == 1 {
			if promptWrapperYes("Do you want to set this directory as the new data directory? (y/n)") {
				if systemctlConfigure {
					// Edit the service file
					editAlgorandServiceFile(paths[0])
				} else {
					affectALGORAND_DATA(paths[0])
				}
				os.Exit(0)
			}

		} else {

			if promptWrapperYes("Do you want to set one of these directories as the new data directory? (y/n)") {

				selectedPath := promptWrapperSelection("Select an Algorand data directory", paths)

				if systemctlConfigure {
					// Edit the service file
					editAlgorandServiceFile(selectedPath)
				} else {
					affectALGORAND_DATA(selectedPath)
				}
				os.Exit(0)
			}
		}
	}

	// Deep search
	if promptWrapperNo("Do you want Algorun to do a deep search for pre-existing Algorand Data directories? (y/n)") {
		fmt.Println("User chose not to search for more pre-existing Algorand Data directories. Exiting...")
		os.Exit(0)
	}

	fmt.Println("Searching for pre-existing Algorand Data directories in HOME directory...")
	paths = deepSearchAlgorandDataDirs()

	if len(paths) == 0 {
		fmt.Println("No Algorand data directories could be found in HOME directory. Are you sure Algorand node has been setup? Please run install command.")
		os.Exit(1)
	}

	fmt.Println("Found Algorand data directories:")
	for _, path := range paths {
		fmt.Println(path)
	}

	// Prompt user to select a directory
	selectedPath := promptWrapperSelection("Select an Algorand data directory", paths)

	if systemctlConfigure {
		editAlgorandServiceFile(selectedPath)
	} else {
		affectALGORAND_DATA(selectedPath)
	}
	os.Exit(0)
}

// Start Algod on your system (the one on your PATH).
func startNode() {
	fmt.Println("Attempting to start Algod...")

	if !isAlgodInstalled() {
		fmt.Println("Algod is not installed. Please run the node install command.")
		os.Exit(1)
	}

	// Check if Algod is already running
	if isAlgodRunning() {
		fmt.Println("Algod is already running.")
		os.Exit(0)
	}

	startAlgodProcess()
}

func isAlgodRunning() bool {
	// Check if Algod is already running
	// This works for systemctl started algorand.service as well as directly started algod
	err := exec.Command("pgrep", "algod").Run()
	return err == nil
}

func startAlgodProcess() {
	// Check if algod is available as a systemctl service
	if checkSystemctlAlgorandServiceCreated() {
		// Algod is available as a systemd service, start it using systemctl

		if !isRunningWithSudo() {
			fmt.Println("This command must be run with super-user priviledges (sudo).")
			os.Exit(1)
		}

		fmt.Println("Starting algod using systemctl...")
		cmd := exec.Command("systemctl", "start", "algorand")
		err := cmd.Run()
		if err != nil {
			fmt.Printf("Failed to start algod service: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Algod is not available as a systemd service, start it directly
		fmt.Println("Starting algod directly...")

		// Check if ALGORAND_DATA environment variable is set
		fmt.Println("Checking if ALGORAND_DATA env var is set...")
		algorandData := os.Getenv("ALGORAND_DATA")

		if !validateAlgorandDataDir(algorandData) {
			fmt.Println("ALGORAND_DATA environment variable is not set or is invalid. Please run node configure and follow the instructions.")
			os.Exit(1)
		}

		fmt.Println("ALGORAND_DATA env var set to valid directory: " + algorandData)

		cmd := exec.Command("algod")
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true,
		}
		err := cmd.Start()
		if err != nil {
			fmt.Printf("Failed to start algod: %v\n", err)
			os.Exit(1)
		}
	}

	// Wait for the process to start
	time.Sleep(5 * time.Second)

	if isAlgodRunning() {
		fmt.Println("Algod is running.")
	} else {
		fmt.Println("Algod failed to start.")
	}
}

// Stop the Algod process on your system.
func stopNode() {
	fmt.Println("Attempting to stop Algod...")

	if !isAlgodRunning() {
		fmt.Println("Algod was not running.")
		os.Exit(0)
	}

	stopAlgodProcess()

	time.Sleep(5 * time.Second)

	if !isAlgodRunning() {
		fmt.Println("Algod is no longer running.")
		os.Exit(0)
	}

	fmt.Println("Failed to stop Algod.")
	os.Exit(1)
}

func stopAlgodProcess() {
	// Check if algod is available as a systemd service
	if checkSystemctlAlgorandServiceCreated() {
		if !isRunningWithSudo() {
			fmt.Println("This command must be run with super-user priviledges (sudo).")
			os.Exit(1)
		}

		// Algod is available as a systemd service, stop it using systemctl
		fmt.Println("Stopping algod using systemctl...")
		cmd := exec.Command("systemctl", "stop", "algorand")
		err := cmd.Run()
		if err != nil {
			fmt.Printf("Failed to stop algod service: %v\n", err)
			cobra.CheckErr(err)
		}
		fmt.Println("Algod service stopped.")
	} else {
		// Algod is not available as a systemd service, stop it directly
		fmt.Println("Stopping algod directly...")
		// Find the process ID of algod
		pid, err := findAlgodPID()
		if err != nil {
			fmt.Printf("Failed to find algod process: %v\n", err)
			cobra.CheckErr(err)
		}

		// Send SIGTERM to the process
		process, err := os.FindProcess(pid)
		if err != nil {
			fmt.Printf("Failed to find process with PID %d: %v\n", pid, err)
			cobra.CheckErr(err)
		}

		err = process.Signal(syscall.SIGTERM)
		if err != nil {
			fmt.Printf("Failed to send SIGTERM to process with PID %d: %v\n", pid, err)
			cobra.CheckErr(err)
		}

		fmt.Println("Sent SIGTERM to algod process.")
	}
}

// Upgrade ALGOD (if installed with package manager).
func upgradeAlgod() {

	if !isAlgodInstalled() {
		fmt.Println("Algod is not installed. Please run the node install command.")
		os.Exit(1)
	}

	// Check if Algod was installed with apt/apt-get
	if checkCmdToolExists("apt") {
		upgradeDebianPackage("apt", "algorand-devtools")
	} else if checkCmdToolExists("apt-get") {
		upgradeDebianPackage("apt-get", "algorand-devtools")
	} else if checkCmdToolExists("dnf") {
		upgradeRpmPackage("dnf", "algorand-devtools")
	} else if checkCmdToolExists("yum") {
		upgradeRpmPackage("yum", "algorand-devtools")
	} else {
		fmt.Println("The *node upgrade* command is currently only available for installations done with an approved package manager. Please use a different method to upgrade.")
		os.Exit(1)
	}
}

// Upgrade a package using the specified Debian package manager
func upgradeDebianPackage(packageManager, packageName string) {
	// Check that we are calling with sudo
	if !isRunningWithSudo() {
		fmt.Println("This command must be run with super-user priviledges (sudo).")
		os.Exit(1)
	}

	// Check if the package is installed and if there are updates available using apt-cache policy
	cmd := exec.Command("apt-cache", "policy", packageName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to check package policy: %v\n", err)
		os.Exit(1)
	}

	outputStr := string(output)
	if strings.Contains(outputStr, "Installed: (none)") {
		fmt.Printf("Package %s is not installed.\n", packageName)
		os.Exit(1)
	}

	installedVersion := extractVersion(outputStr, "Installed:")
	candidateVersion := extractVersion(outputStr, "Candidate:")

	if installedVersion == candidateVersion {
		fmt.Printf("Package %s is installed (v%s) and up-to-date with latest (v%s).\n", packageName, installedVersion, candidateVersion)
		os.Exit(0)
	}

	fmt.Printf("Package %s is installed (v%s) and has updates available (v%s).\n", packageName, installedVersion, candidateVersion)

	// Update the package list
	fmt.Println("Updating package list...")
	cmd = exec.Command(packageManager, "update")
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Failed to update package list: %v\n", err)
		os.Exit(1)
	}

	// Upgrade the package
	fmt.Printf("Upgrading package %s...\n", packageName)
	cmd = exec.Command(packageManager, "install", "--only-upgrade", "-y", packageName)
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Failed to upgrade package %s: %v\n", packageName, err)
		os.Exit(1)
	}

	fmt.Printf("Package %s upgraded successfully.\n", packageName)
	os.Exit(0)
}

// Upgrade a package using the specified RPM package manager
func upgradeRpmPackage(packageManager, packageName string) {
	// Check that we are calling with sudo
	if !isRunningWithSudo() {
		fmt.Println("This command must be run with sudo.")
		os.Exit(1)
	}

	// Attempt to upgrade the package directly
	fmt.Printf("Upgrading package %s...\n", packageName)
	cmd := exec.Command(packageManager, "update", "-y", packageName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to upgrade package %s: %v\n", packageName, err)
		os.Exit(1)
	}

	outputStr := string(output)
	if strings.Contains(outputStr, "Nothing to do") {
		fmt.Printf("Package %s is already up-to-date.\n", packageName)
		os.Exit(0)
	} else {
		fmt.Println(outputStr)
		fmt.Printf("Package %s upgraded successfully.\n", packageName)
		os.Exit(0)
	}
}
