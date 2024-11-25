package node

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Algod",
	Long:  "Start Algod on your system (the one on your PATH).",
	Run: func(cmd *cobra.Command, args []string) {
		startNode()
	},
}

// Start Algod on your system (the one on your PATH).
func startNode() {
	fmt.Println("Attempting to start Algod...")

	if !isAlgodInstalled() {
		fmt.Println("Algod is not installed. Please run the *node install* command.")
		os.Exit(1)
	}

	// Check if Algod is already running
	if isAlgodRunning() {
		fmt.Println("Algod is already running.")
		os.Exit(0)
	}

	startAlgodProcess()
}

func startAlgodProcess() {

	if !isRunningWithSudo() {
		fmt.Println("This command must be run with super-user priviledges (sudo).")
		os.Exit(1)
	}

	// Check if algod is available as a system service
	if checkAlgorandServiceCreated() {
		// Algod is available as a service

		switch runtime.GOOS {
		case "linux":
			startSystemdAlgorandService()
		case "darwin":
			startLaunchdAlgorandService()
		default: // Unsupported OS
			fmt.Println("Unsupported OS.")
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

// Linux uses systemd
func startSystemdAlgorandService() {
	fmt.Println("Starting algod using systemctl...")
	cmd := exec.Command("systemctl", "start", "algorand")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Failed to start algod service: %v\n", err)
		os.Exit(1)
	}
}

// MacOS uses launchd instead of systemd
func startLaunchdAlgorandService() {
	fmt.Println("Starting algod using launchctl...")
	cmd := exec.Command("launchctl", "load", "/Library/LaunchDaemons/com.algorand.algod.plist")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Failed to start algod service: %v\n", err)
		os.Exit(1)
	}
}
