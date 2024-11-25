package node

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

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

	fmt.Println("Algod is installed. Uninstalling...")

	// Check if Algod is running
	if isAlgodRunning() {
		fmt.Println("Algod is running. Please run *node stop*.")
		os.Exit(1)
	}

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

func unInstallNodeLinux() {}
