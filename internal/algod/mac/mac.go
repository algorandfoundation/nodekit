package mac

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/algorandfoundation/algorun-tui/internal/system"
	"github.com/spf13/cobra"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

// IsService check if Algorand service has been created with launchd (macOS)
// Note that it needs to be run in super-user privilege mode to
// be able to view the root level services.
func IsService() bool {
	_, err := system.Run([]string{"launchctl", "list", "com.algorand.algod"})
	return err == nil
}

// Install sets up Algod on macOS using Homebrew,
// configures necessary directories, and ensures it
// runs as a background service.
func Install() error {
	fmt.Println("Installing Algod on macOS...")

	// Homebrew is our package manager of choice
	if !system.CmdExists("brew") {
		return fmt.Errorf("could not find Homebrew installed. Please install Homebrew and try again")
	}

	err := system.RunAll(system.CmdsList{
		{"brew", "tap", "algorandfoundation/homebrew-node"},
		{"brew", "install", "algorand"},
		{"brew", "--prefix", "algorand", "--installed"},
	})
	if err != nil {
		return err
	}

	// Handle data directory and genesis.json file
	handleDataDirMac()

	path, err := os.Executable()
	if err != nil {
		return err
	}
	// Create and load the launchd service
	_, err = system.Run([]string{"sudo", path, "configure", "service"})
	if err != nil {
		return fmt.Errorf("failed to create and load launchd service: %v\n", err)
	}

	// Ensure Homebrew bin directory is in the PATH
	// So that brew installed algorand binaries can be found
	ensureHomebrewPathInEnv()

	if !IsService() {
		return fmt.Errorf("algod unexpectedly NOT in path. Installation failed")
	}

	fmt.Println(`Installed Algorand (Algod) with Homebrew.
Algod is running in the background as a system-level service.
	`)

	return nil
}

// Uninstall removes the Algorand application from the system using Homebrew if it is installed.
func Uninstall() error {
	if !system.CmdExists("brew") {
		return errors.New("homebrew is not installed")
	}
	return exec.Command("brew", "uninstall", "algorand").Run()
}

// Upgrade updates the installed Algorand package using Homebrew if it's available and properly configured.
func Upgrade() error {
	if !system.CmdExists("brew") {
		return errors.New("homebrew is not installed")
	}
	// Check if algorand is installed with Homebrew
	checkCmdArgs := system.CmdsList{{"brew", "--prefix", "algorand", "--installed"}}
	if system.IsSudo() {
		checkCmdArgs = checkCmdArgs.Su(os.Getenv("SUDO_USER"))
	}
	err := system.RunAll(checkCmdArgs)
	if err != nil {
		return err
	}
	// Upgrade algorand
	upgradeCmdArgs := system.CmdsList{{"brew", "upgrade", "algorand", "--formula"}}
	if system.IsSudo() {
		upgradeCmdArgs = upgradeCmdArgs.Su(os.Getenv("SUDO_USER"))
	}
	return system.RunAll(upgradeCmdArgs)
}

// Start algorand with launchd
func Start() error {
	return exec.Command("launchctl", "load", "/Library/LaunchDaemons/com.algorand.algod.plist").Run()
}

// Stop shuts down the Algorand algod system process using the launchctl bootout command.
// Returns an error if the operation fails.
func Stop() error {
	return exec.Command("launchctl", "bootout", "system/com.algorand.algod").Run()
}

// UpdateService updates the Algorand launchd service with
// a new data directory path and reloads the service configuration.
func UpdateService(dataDirectoryPath string) error {

	algodPath, err := exec.LookPath("algod")
	if err != nil {
		fmt.Printf("Failed to find algod binary: %v\n", err)
		os.Exit(1)
	}

	overwriteFilePath := "/Library/LaunchDaemons/com.algorand.algod.plist"

	overwriteTemplate := `<?xml version="1.0" encoding="UTF-8"?>
	<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
	<plist version="1.0">
	<dict>
					<key>Label</key>
					<string>com.algorand.algod</string>
					<key>ProgramArguments</key>
					<array>
													<string>{{.AlgodPath}}</string>
													<string>-d</string>
													<string>{{.DataDirectoryPath}}</string>
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
	</plist>`

	// Data to fill the template
	data := map[string]string{
		"AlgodPath":         algodPath,
		"DataDirectoryPath": dataDirectoryPath,
	}

	// Parse and execute the template
	tmpl, err := template.New("override").Parse(overwriteTemplate)
	if err != nil {
		fmt.Printf("Failed to parse template: %v\n", err)
		os.Exit(1)
	}

	var overwriteContent bytes.Buffer
	err = tmpl.Execute(&overwriteContent, data)
	if err != nil {
		fmt.Printf("Failed to execute template: %v\n", err)
		os.Exit(1)
	}

	// Write the override content to the file
	err = os.WriteFile(overwriteFilePath, overwriteContent.Bytes(), 0644)
	if err != nil {
		fmt.Printf("Failed to write override file: %v\n", err)
		os.Exit(1)
	}

	// Boot out the launchd service (just in case - it should be off)
	cmd := exec.Command("launchctl", "bootout", "system", overwriteFilePath)
	err = cmd.Run()
	if err != nil {
		if !strings.Contains(err.Error(), "No such process") {
			fmt.Printf("Failed to bootout launchd service: %v\n", err)
			os.Exit(1)
		}
	}

	// Load the launchd service
	cmd = exec.Command("launchctl", "load", overwriteFilePath)
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Failed to load launchd service: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Launchd service updated and reloaded successfully.")
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

func EnsureService() error {
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

	return nil
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
