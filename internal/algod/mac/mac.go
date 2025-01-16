package mac

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/algorandfoundation/nodekit/internal/algod/utils"
	"github.com/algorandfoundation/nodekit/internal/system"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
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
func IsService() bool {
	_, err := system.Run([]string{"sudo", "launchctl", "list", "com.algorand.algod"})
	return err == nil
}

// Install sets up Algod on macOS using Homebrew,
// configures necessary directories, and ensures it
// runs as a background service.
func Install() error {
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
	err = handleDataDirMac()
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

	if !IsService() {
		return fmt.Errorf("algod is not a service")
	}

	log.Info("Installed Algorand (Algod) with Homebrew ")

	return nil
}

// Uninstall removes the Algorand application from the system using Homebrew if it is installed.
func Uninstall(force bool) error {
	if force {
		if system.IsCmdRunning("algod") {
			err := Stop(force)
			if err != nil {
				return err
			}
		}
	}

	cmds := system.CmdsList{}
	if IsService() {
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
func Upgrade(force bool) error {
	if !system.CmdExists("brew") {
		return errors.New("homebrew is not installed")
	}
	err := system.RunAll(system.CmdsList{
		{"brew", "--prefix", "algorand", "--installed"},
		{"brew", "update"},
		{"brew", "upgrade", "algorand", "--formula"},
	})
	if err != nil {
		return err
	}
	err = Stop(false)
	if err != nil {
		return err
	}
	return Start(false)
}

// Start algorand with launchd
func Start(force bool) error {
	log.Debug("Attempting to start algorand with launchd")
	//if !IsService() && !force {
	//	return fmt.Errorf(MustBeServiceMsg)
	//}
	return system.RunAll(system.CmdsList{
		{"sudo", "launchctl", "start", "com.algorand.algod"},
	})
}

// Stop shuts down the Algorand algod system process using the launchctl bootout command.
// Returns an error if the operation fails.
func Stop(force bool) error {
	if !IsService() && !force {
		return fmt.Errorf(MustBeServiceMsg)
	}

	return system.RunAll(system.CmdsList{
		{"sudo", "launchctl", "stop", "com.algorand.algod"},
	})
}

// UpdateService updates the Algorand launchd service with
// a new data directory path and reloads the service configuration.
// TODO: Deduplicate this method, redundant version of EnsureService.
func UpdateService(dataDirectoryPath string) error {

	algodPath, err := exec.LookPath("algod")
	if err != nil {
		log.Info("Failed to find algod binary: %v\n", err)
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
		log.Info("Failed to parse template: %v\n", err)
		os.Exit(1)
	}

	var overwriteContent bytes.Buffer
	err = tmpl.Execute(&overwriteContent, data)
	if err != nil {
		log.Info("Failed to execute template: %v\n", err)
		os.Exit(1)
	}

	// Write the override content to the file
	err = os.WriteFile(overwriteFilePath, overwriteContent.Bytes(), 0644)
	if err != nil {
		log.Info("Failed to write override file: %v\n", err)
		os.Exit(1)
	}

	// Boot out the launchd service (just in case - it should be off)
	cmd := exec.Command("launchctl", "bootout", "system", overwriteFilePath)
	err = cmd.Run()
	if err != nil {
		if !strings.Contains(err.Error(), "No such process") {
			log.Info("Failed to bootout launchd service: %v\n", err)
			os.Exit(1)
		}
	}

	// Load the launchd service
	cmd = exec.Command("launchctl", "load", overwriteFilePath)
	err = cmd.Run()
	if err != nil {
		log.Info("Failed to load launchd service: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Launchd service updated and reloaded successfully.")
	return nil
}

// handleDataDirMac ensures the necessary Algorand data directory and mainnet genesis.json file exist on macOS.
// TODO move to configure as a generic
func handleDataDirMac() error {
	// Ensure the ~/.algorand directory exists
	algorandDir := filepath.Join(os.Getenv("HOME"), ".algorand")
	if err := os.MkdirAll(algorandDir, 0755); err != nil {
		return err
	}

	// Check if genesis.json file exists in ~/.algorand
	// TODO: replace with algocfg or goal templates
	genesisFilePath := filepath.Join(os.Getenv("HOME"), ".algorand", "genesis.json")
	_, err := os.Stat(genesisFilePath)
	if !os.IsNotExist(err) {
		return nil
	}

	log.Info("Downloading mainnet genesis.json file to ~/.algorand/genesis.json")

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

// EnsureService ensures the `algod` service is properly configured and running as a background service on macOS.
// It checks for the existence of the `algod` binary, creates a launchd plist file, and loads it using `launchctl`.
// Returns an error if the binary is not found, or if any system command fails.
func EnsureService() error {
	log.Debug("Ensuring Algorand service is running")
	path, err := exec.LookPath("algod")
	if err != nil {
		log.Error("algod does not exist in path")
		return err
	}
	// Define the launchd plist content
	plistContent := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.algorand.algod</string>
	<key>ProgramArguments</key>
	<array>
			<string>%s</string>
			<string>-d</string>
			<string>%s/.algorand</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
    <key>Debug</key>
    <true/>
	<key>StandardOutPath</key>
	<string>/tmp/algod.out</string>
	<key>StandardErrorPath</key>
	<string>/tmp/algod.err</string>
</dict>
</plist>`, path, os.Getenv("HOME"))

	// Write the plist content to a file
	plistPath := "/Library/LaunchDaemons/com.algorand.algod.plist"
	err = os.MkdirAll(filepath.Dir(plistPath), 0755)
	if err != nil {
		log.Info("Failed to create LaunchDaemons directory: %v\n", err)
		cobra.CheckErr(err)
	}

	err = os.WriteFile(plistPath, []byte(plistContent), 0644)
	if err != nil {
		log.Info("Failed to write plist file: %v\n", err)
		cobra.CheckErr(err)
	}
	return system.RunAll(system.CmdsList{
		{"launchctl", "load", plistPath},
		{"launchctl", "list", "com.algorand.algod"},
	})
}
