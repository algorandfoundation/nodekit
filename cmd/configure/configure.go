package configure

import (
	"bytes"
	"fmt"
	"github.com/algorandfoundation/algorun-tui/internal/algod"
	"github.com/algorandfoundation/algorun-tui/internal/algod/utils"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:          "configure",
	Short:        "Configure Algod",
	Long:         "Configure Algod settings",
	SilenceUsage: true,
	//PersistentPreRun:
	//RunE: func(cmd *cobra.Command, args []string) error {
	//	return configureNode()
	//},
}

func init() {
	Cmd.AddCommand(serviceCmd)
}

const ConfigureRunningErrorMsg = "algorand is currently running. Please stop the node with *node stop* before configuring"

// TODO: configure not just data directory but algod path
func configureNode() error {
	var systemServiceConfigure bool

	if algod.IsRunning() {
		return fmt.Errorf(ConfigureRunningErrorMsg)
	}

	// Check systemctl first
	if algod.IsService() {
		if promptWrapperYes("Algorand is installed as a service. Do you wish to edit the service file to change the data directory? (y/n)") {
			// Edit the service file with the user's new data directory
			systemServiceConfigure = true
		} else {
			fmt.Println("Exiting...")
			os.Exit(0)
		}
	}

	// At the end, instead of affectALGORAND_DATA, we'll edit the systemctl algorand.service file
	// i.e., overwrite /etc/systemd/system/algorand.service.d/override.conf
	// ExecStart and Description will be changed to reflect the new data directory
	//

	if !systemServiceConfigure {
		fmt.Println("Configuring Data directory for algod started through Algorun...")
	}

	algorandData := os.Getenv("ALGORAND_DATA")

	// Check if ALGORAND_DATA environment variable is set
	if algorandData != "" {
		fmt.Println("ALGORAND_DATA environment variable is set to: " + algorandData)
		fmt.Println("Inspecting the set data directory...")

		if validateAlgorandDataDir(algorandData) {
			fmt.Println("Found valid Algorand Data Directory: " + algorandData)

			if systemServiceConfigure {
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

				if systemServiceConfigure {
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
	paths := utils.GetKnownDataPaths()

	if len(paths) != 0 {

		fmt.Println("Quick check found the following potential data directories:")
		for _, path := range paths {
			fmt.Println("✔ " + path)
		}

		if len(paths) == 1 {
			if promptWrapperYes("Do you want to set this directory as the new data directory? (y/n)") {
				if systemServiceConfigure {
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

				if systemServiceConfigure {
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

	if systemServiceConfigure {
		editAlgorandServiceFile(selectedPath)
	} else {
		affectALGORAND_DATA(selectedPath)
	}
	return nil
}

func editAlgorandServiceFile(dataDirectoryPath string) {
	switch runtime.GOOS {
	case "linux":
		editSystemdAlgorandServiceFile(dataDirectoryPath)
	case "darwin":
		editLaunchdAlgorandServiceFile(dataDirectoryPath)
	default:
		fmt.Println("Unsupported operating system.")
	}
}

func editLaunchdAlgorandServiceFile(dataDirectoryPath string) {

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
}

// Update the algorand.service file
func editSystemdAlgorandServiceFile(dataDirectoryPath string) {

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

	fmt.Println("Algorand service file updated successfully.")
}
