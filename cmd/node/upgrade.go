package node

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade Algod",
	Long:  "Upgrade Algod (if installed with package manager).",
	Run: func(cmd *cobra.Command, args []string) {
		upgradeAlgod()
	},
}

// Upgrade ALGOD (if installed with package manager).
func upgradeAlgod() {
	if !isAlgodInstalled() {
		fmt.Println("Algod is not installed.")
		os.Exit(1)
	}

	switch runtime.GOOS {
	case "darwin":
		if checkCmdToolExists("brew") {
			upgradeBrewAlgorand()
		} else {
			fmt.Println("Homebrew is not installed. Please install Homebrew and try again.")
			os.Exit(1)
		}
	case "linux":
		// Check if Algod was installed with apt/apt-get, dnf, or yum
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
	default:
		fmt.Println("Unsupported operating system. The *node upgrade* command is only available for macOS and Linux.")
		os.Exit(1)
	}
}

func upgradeBrewAlgorand() {
	fmt.Println("Upgrading Algod using Homebrew...")

	var prefixCommand []string

	// Brew cannot be run with sudo, so we need to run the commands as the original user.
	// This checks if the user has ran this command with super-user privileges, and if so
	// counteracts it by running the commands as the original user.
	if isRunningWithSudo() {
		originalUser := os.Getenv("SUDO_USER")
		prefixCommand = []string{"sudo", "-u", originalUser}
	} else {
		prefixCommand = []string{}
	}

	// Check if algorand is installed with Homebrew
	checkCmdArgs := append(prefixCommand, "brew", "--prefix", "algorand", "--installed")
	fmt.Println("Running command:", strings.Join(checkCmdArgs, " "))
	checkCmd := exec.Command(checkCmdArgs[0], checkCmdArgs[1:]...)
	checkCmd.Stdout = os.Stdout
	checkCmd.Stderr = os.Stderr
	if err := checkCmd.Run(); err != nil {
		fmt.Println("Algorand is not installed with Homebrew.")
		os.Exit(1)
	}

	// Upgrade algorand
	upgradeCmdArgs := append(prefixCommand, "brew", "upgrade", "algorand", "--formula")
	fmt.Println("Running command:", strings.Join(upgradeCmdArgs, " "))
	upgradeCmd := exec.Command(upgradeCmdArgs[0], upgradeCmdArgs[1:]...)
	upgradeCmd.Stdout = os.Stdout
	upgradeCmd.Stderr = os.Stderr
	if err := upgradeCmd.Run(); err != nil {
		fmt.Printf("Failed to upgrade Algorand: %v\n", err)
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
