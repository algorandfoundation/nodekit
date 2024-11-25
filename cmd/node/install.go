package node

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Algorand node (Algod)",
	Long:  "Install Algorand node (Algod) and other binaries on your system",
	Run: func(cmd *cobra.Command, args []string) {
		installNode()
	},
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

	// Check that we are calling with sudo
	if !isRunningWithSudo() {
		fmt.Println("This command must be run with super-user privileges (sudo).")
		os.Exit(1)
	}

	// Homebrew is our package manager of choice
	if !checkCmdToolExists("brew") {
		fmt.Println("Could not find Homebrew installed. Please install Homebrew and try again.")
		os.Exit(1)
	}

	originalUser := os.Getenv("SUDO_USER")

	// Run Homebrew commands as the original user without sudo
	if err := runHomebrewInstallCommandsAsUser(originalUser); err != nil {
		fmt.Printf("Homebrew commands failed: %v\n", err)
		os.Exit(1)
	}

	// Handle data directory and genesis.json file
	handleDataDirMac()

	// Create and load the launchd service
	createAndLoadLaunchdService()

	// Ensure Homebrew bin directory is in the PATH
	// So that brew installed algorand binaries can be found
	ensureHomebrewPathInEnv()

	if !isAlgodInstalled() {
		fmt.Println("algod unexpectedly NOT in path. Installation failed.")
		os.Exit(1)
	}

	fmt.Println(`Installed Algorand (Algod) with Homebrew.
Algod is running in the background as a system-level service.
	`)
	os.Exit(0)
}

func runHomebrewInstallCommandsAsUser(user string) error {
	homebrewCmds := [][]string{
		{"brew", "tap", "HashMapsData2Value/homebrew-tap"},
		{"brew", "install", "algorand"},
		{"brew", "--prefix", "algorand", "--installed"},
	}

	for _, cmdArgs := range homebrewCmds {
		fmt.Println("Running command:", strings.Join(cmdArgs, " "))
		cmd := exec.Command("sudo", append([]string{"-u", user}, cmdArgs...)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command failed: %s\nError: %v", strings.Join(cmdArgs, " "), err)
		}
	}
	return nil
}

func handleDataDirMac() {
	// Ensure the ~/.algorand directory exists
	algorandDir := filepath.Join(os.Getenv("HOME"), ".algorand")
	if err := os.MkdirAll(algorandDir, 0755); err != nil {
		fmt.Printf("Failed to create directory %s: %v\n", algorandDir, err)
		os.Exit(1)
	}

	// Check if genesis.json file exists in ~/.algorand
	genesisFilePath := filepath.Join(os.Getenv("HOME"), ".algorand", "genesis.json")
	if _, err := os.Stat(genesisFilePath); os.IsNotExist(err) {
		fmt.Println("genesis.json file does not exist. Downloading...")

		// Download the genesis.json file
		resp, err := http.Get("https://raw.githubusercontent.com/algorand/go-algorand/db7f1627e4919b05aef5392504e48b93a90a0146/installer/genesis/mainnet/genesis.json")
		if err != nil {
			fmt.Printf("Failed to download genesis.json: %v\n", err)
			cobra.CheckErr(err)
		}
		defer resp.Body.Close()

		// Create the file
		out, err := os.Create(genesisFilePath)
		if err != nil {
			fmt.Printf("Failed to create genesis.json file: %v\n", err)
			cobra.CheckErr(err)
		}
		defer out.Close()

		// Write the content to the file
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			fmt.Printf("Failed to save genesis.json file: %v\n", err)
			cobra.CheckErr(err)
		}

		fmt.Println("mainnet genesis.json file downloaded successfully.")
	}

}

func createAndLoadLaunchdService() {
	// Get the prefix path for Algorand
	cmd := exec.Command("brew", "--prefix", "algorand")
	algorandPrefix, err := cmd.Output()
	if err != nil {
		fmt.Printf("Failed to get Algorand prefix: %v\n", err)
		cobra.CheckErr(err)
	}
	algorandPrefixPath := strings.TrimSpace(string(algorandPrefix))

	// Define the launchd plist content
	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.algorand.algod</string>
	<key>ProgramArguments</key>
	<array>
			<string>%s/bin/algod</string>
			<string>-d</string>
			<string>%s/.algorand</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
	<key>StandardOutPath</key>
	<string>/tmp/algod.out</string>
	<key>StandardErrorPath</key>
	<string>/tmp/algod.err</string>
</dict>
</plist>`, algorandPrefixPath, os.Getenv("HOME"))

	// Write the plist content to a file
	plistPath := "/Library/LaunchDaemons/com.algorand.algod.plist"
	err = os.MkdirAll(filepath.Dir(plistPath), 0755)
	if err != nil {
		fmt.Printf("Failed to create LaunchDaemons directory: %v\n", err)
		cobra.CheckErr(err)
	}

	err = os.WriteFile(plistPath, []byte(plistContent), 0644)
	if err != nil {
		fmt.Printf("Failed to write plist file: %v\n", err)
		cobra.CheckErr(err)
	}

	// Load the launchd service
	cmd = exec.Command("launchctl", "load", plistPath)
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Failed to load launchd service: %v\n", err)
		cobra.CheckErr(err)
	}

	// Check if the service is running
	cmd = exec.Command("launchctl", "list", "com.algorand.algod")
	err = cmd.Run()
	if err != nil {
		fmt.Printf("launchd service is not running: %v\n", err)
		cobra.CheckErr(err)
	}

	fmt.Println("Launchd service created and loaded successfully.")
}

// Ensure that Homebrew bin directory is in the PATH so that Algorand binaries can be found
func ensureHomebrewPathInEnv() {
	homebrewPrefix := os.Getenv("HOMEBREW_PREFIX")
	homebrewCellar := os.Getenv("HOMEBREW_CELLAR")
	homebrewRepository := os.Getenv("HOMEBREW_REPOSITORY")

	if homebrewPrefix == "" || homebrewCellar == "" || homebrewRepository == "" {
		fmt.Println("Homebrew environment variables are not set. Running brew shellenv...")

		cmd := exec.Command("brew", "shellenv")
		output, err := cmd.Output()
		if err != nil {
			fmt.Printf("Failed to get Homebrew environment: %v\n", err)
			return
		}

		envVars := strings.Split(string(output), "\n")
		for _, envVar := range envVars {
			if envVar != "" {
				fmt.Println("Setting environment variable:", envVar)
				os.Setenv(strings.Split(envVar, "=")[0], strings.Trim(strings.Split(envVar, "=")[1], `"`))
			}
		}

		// Append brew shellenv output to .zshrc
		zshrcPath := filepath.Join(os.Getenv("HOME"), ".zshrc")
		f, err := os.OpenFile(zshrcPath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("Failed to open .zshrc: %v\n", err)
			fmt.Printf("Are you running a terminal other than zsh?")
			fmt.Printf("Please run brew shellenv and add the output to your shell's configuration file.")
			return
		}
		defer f.Close()

		if _, err := f.WriteString("\n# Inserted by Algorun\n# Homebrew environment variables\n" + string(output)); err != nil {
			fmt.Printf("Failed to write to .zshrc: %v\n", err)
		}
	}
}
