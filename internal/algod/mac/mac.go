package mac

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/algorandfoundation/nodekit/internal/algod/fallback"
	"github.com/algorandfoundation/nodekit/internal/algod/utils"
	"github.com/algorandfoundation/nodekit/internal/system"
	"github.com/charmbracelet/log"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
)

// MustBeServiceMsg is an error message indicating that a service must be installed to manage it.
const MustBeServiceMsg = "service must be installed to be able to manage it"

// HomeBrewNotFoundMsg is the error message returned when Homebrew is not detected on the system during execution.
const HomeBrewNotFoundMsg = "homebrew is not installed. please install Homebrew and try again"

// IsService check if Algorand service has been created with launchd (macOS)
// Note that it needs to be run in super-user privilege mode to
// be able to view the root level services.
func IsService(dataDir string) bool {
	_, err := system.Run([]string{"sudo", "launchctl", "list", "com.algorand.algod"})
	return err == nil
}

// Install sets up Algod on macOS using Homebrew,
// configures necessary directories, and ensures it
// runs as a background service.
func Install(dataDir string, network string, force bool) error {
	if dataDir == "" {
		log.Info("Installing Algod on macOS...")

		// Homebrew is our package manager of choice
		if !system.CmdExists("brew") {
			return errors.New(HomeBrewNotFoundMsg)
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
		err = fallback.SetNetwork(dataDir, network, false)
		if err != nil {
			return err
		}

		path, err := os.Executable()
		if err != nil {
			return err
		}

		// Create and load the launchd service
		// TODO: find a clever way to avoid this or make sudo persist for the second call
		err = system.RunAll(system.CmdsList{{"sudo", path, "configure", "service"}})
		if err != nil {
			return err
		}

		if !IsService("") {
			return fmt.Errorf("algod is not a service")
		}

		log.Info("Installed Algorand (Algod) with Homebrew ")
	}

	return fallback.Install("", dataDir, false)
}

// Uninstall removes the Algorand application from the system using Homebrew if it is installed.
func Uninstall(dataDir string, force bool) error {
	if force {
		if system.IsCmdRunning("algod") {
			err := Stop(dataDir)
			if err != nil {
				return err
			}
		}
	}

	cmds := system.CmdsList{}
	if IsService("") {
		cmds = append(cmds, []string{"sudo", "launchctl", "unload", "/Library/LaunchDaemons/com.algorand.algod.plist"})
	}

	if !system.CmdExists("brew") && !force {
		return errors.New("homebrew is not installed")
	} else {
		cmds = append(cmds, []string{"brew", "uninstall", "algorand"})
	}

	if force {
		cmds = append(cmds, []string{"sudo", "rm", "-rf", strings.Join(utils.GetKnownDataPaths(), " ")})
		cmds = append(cmds, []string{"sudo", "rm", "-rf", "/Library/LaunchDaemons/com.algorand.algod.plist"})
	}

	return system.RunAll(cmds)
}

// Upgrade updates the installed Algorand package using Homebrew if it's available and properly configured.
func Upgrade() error {
	if !system.CmdExists("brew") {
		return errors.New("homebrew is not installed")
	}

	return system.RunAll(system.CmdsList{
		{"brew", "--prefix", "algorand", "--installed"},
		{"brew", "upgrade", "algorand", "--formula"},
	})
}

// Start algorand with launchd or optionally use the fallback
func Start(dataDir string) error {
	//if dataDir == "" {
	log.Debug("Attempting to start algorand with launchd")
	return system.RunAll(system.CmdsList{
		{"sudo", "launchctl", "start", "com.algorand.algod"},
	})
	//}
	//return fallback.Start(dataDir)
}

// Stop shuts down the Algorand algod system process using the launchctl bootout command.
// Returns an error if the operation fails.
func Stop(dataDir string) error {
	if dataDir == "" {
		log.Debug("Attempting to stop algorand with launchd")

		return system.RunAll(system.CmdsList{
			{"sudo", "launchctl", "stop", "com.algorand.algod"},
		})
	}
	return fallback.Stop(dataDir)
}

// UpdateService updates the Algorand launchd service with
// a new data directory path and reloads the service configuration.
func UpdateService(dataDirectoryPath string) error {
	algodPath, err := exec.LookPath("algod")
	if err != nil {
		return err
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
		return err
	}

	var overwriteContent bytes.Buffer
	err = tmpl.Execute(&overwriteContent, data)
	if err != nil {
		return err
	}

	// Write the override content to the file
	err = os.WriteFile(overwriteFilePath, overwriteContent.Bytes(), 0644)
	if err != nil {
		return err
	}

	return system.RunAll(system.CmdsList{
		{"launchctl", "load", overwriteFilePath},
		{"launchctl", "list", "com.algorand.algod"},
	})
}

// handleDataDirMac ensures the necessary Algorand data directory and mainnet genesis.json file exist on macOS.
// TODO move to configure as a generic for both linux/mac
func handleDataDirMac(dataDir string) error {
	// Ensure the ~/.algorand directory exists
	var algorandDir string
	if dataDir != "" {
		algorandDir = dataDir
	} else {
		algorandDir = filepath.Join(os.Getenv("HOME"), ".algorand")
	}
	if err := os.MkdirAll(algorandDir, 0755); err != nil {
		return err
	}

	// Check if genesis.json file exists
	// TODO: replace with algocfg or goal templates
	genesisFilePath := filepath.Join(algorandDir, "genesis.json")
	_, err := os.Stat(genesisFilePath)
	// Return if a genesis file is already configured
	if !os.IsNotExist(err) {
		return nil
	}

	log.Info(fmt.Sprintf("Downloading mainnet genesis.json file to %s/genesis.json", algorandDir))

	// Download the genesis.json file
	resp, err := http.Get("https://raw.githubusercontent.com/algorand/go-algorand/db7f1627e4919b05aef5392504e48b93a90a0146/installer/genesis/mainnet/genesis.json")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(genesisFilePath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the content to the file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	log.Info("mainnet genesis.json file downloaded successfully.")
	return nil
}
