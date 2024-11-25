package node

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall Algorand node (Algod)",
	Long:  "Uninstall Algorand node (Algod) and other binaries on your system installed by this tool.",
	Run: func(cmd *cobra.Command, args []string) {
		unInstallNode()
	},
}

// Uninstall Algorand node (Algod) and other binaries on your system
func unInstallNode() {

	if !isRunningWithSudo() {
		fmt.Println("This command must be run with super-user priviledges (sudo).")
		os.Exit(1)
	}

	fmt.Println("Checking if Algod is installed...")

	// Check if Algod is installed
	if !isAlgodInstalled() {
		fmt.Println("Algod is not installed.")
		os.Exit(0)
	}

	// Check if Algod is running
	if isAlgodRunning() {
		fmt.Println("Algod is running. Please run *node stop* first to stop it.")
		os.Exit(1)
	}

	fmt.Println("Algod is installed. Proceeding...")

	// Uninstall Algod based on OS
	switch runtime.GOOS {
	case "linux":
		unInstallNodeLinux()
	case "darwin":
		unInstallNodeMac()
	default:
		panic("Unsupported OS: " + runtime.GOOS)
	}

	os.Exit(0)
}

func unInstallNodeMac() {
	fmt.Println("Uninstalling Algod on macOS...")

	// Homebrew is our package manager of choice
	if !checkCmdToolExists("brew") {
		fmt.Println("Could not find Homebrew installed. Did you install Algod some other way?.")
		os.Exit(1)
	}

	user := os.Getenv("SUDO_USER")

	// Run the brew uninstall command as the original user without sudo
	cmd := exec.Command("sudo", "-u", user, "brew", "uninstall", "algorand", "--formula")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to uninstall Algorand: %v\n", err)
		fmt.Printf("Output: %s\n", string(output))
		os.Exit(1)
	}

	fmt.Printf("Output: %s\n", string(output))

	// Calling brew uninstall algorand without sudo user privileges
	cmd = exec.Command("sudo", "-u", user, "brew", "--prefix", "algorand", "--installed")
	err = cmd.Run()
	if err == nil {
		fmt.Println("Algorand uninstall failed.")
		os.Exit(1)
	}

	// Delete the launchd plist file
	plistPath := "/Library/LaunchDaemons/com.algorand.algod.plist"
	err = os.Remove(plistPath)
	if err != nil {
		fmt.Printf("Failed to delete plist file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Algorand uninstalled successfully.")
}

func unInstallNodeLinux() {

	var unInstallCmds [][]string

	if checkCmdToolExists("apt") { // On Ubuntu and Debian we use the apt package manager
		fmt.Println("Using apt package manager")
		unInstallCmds = [][]string{
			{"apt", "remove", "algorand-devtools", "-y"},
			{"apt", "autoremove", "-y"},
		}
	} else if checkCmdToolExists("apt-get") {
		fmt.Println("Using apt-get package manager")
		unInstallCmds = [][]string{
			{"apt-get", "remove", "algorand-devtools", "-y"},
			{"apt-get", "autoremove", "-y"},
		}
	} else if checkCmdToolExists("dnf") { // On Fedora and CentOs8 there's the dnf package manager
		fmt.Println("Using dnf package manager")
		unInstallCmds = [][]string{
			{"dnf", "remove", "algorand-devtools", "-y"},
		}
	} else if checkCmdToolExists("yum") { // On CentOs7 we use the yum package manager
		fmt.Println("Using yum package manager")
		unInstallCmds = [][]string{
			{"yum", "remove", "algorand-devtools", "-y"},
		}
	} else {
		fmt.Println("Could not find a package manager to uninstall Algorand.")
		os.Exit(1)
	}

	// Commands to clear systemd algorand.service and any other files, like the configuration override
	unInstallCmds = append(unInstallCmds, []string{"bash", "-c", "sudo rm -rf /etc/systemd/system/algorand*"})
	unInstallCmds = append(unInstallCmds, []string{"systemctl", "daemon-reload"})

	// Run each installation command
	for _, cmdArgs := range unInstallCmds {
		fmt.Println("Running command:", strings.Join(cmdArgs, " "))
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Command failed: %s\nOutput: %s\nError: %v\n", strings.Join(cmdArgs, " "), output, err)
			cobra.CheckErr(err)
		}
	}

	// Check the status of the algorand service
	cmd := exec.Command("systemctl", "status", "algorand")
	output, err := cmd.CombinedOutput()
	if err != nil && strings.Contains(string(output), "Unit algorand.service could not be found.") {
		fmt.Println("Algorand service has been successfully removed.")
	} else {
		fmt.Printf("Failed to verify Algorand service uninstallation: %v\n", err)
		fmt.Printf("Output: %s\n", string(output))
		os.Exit(1)
	}

	fmt.Println("Algorand successfully uninstalled.")
}
