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

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop Algod",
	Long:  "Stop the Algod process on your system.",
	Run: func(cmd *cobra.Command, args []string) {
		stopNode()
	},
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

	if !isRunningWithSudo() {
		fmt.Println("This command must be run with super-user priviledges (sudo).")
		os.Exit(1)
	}

	// Check if algod is available as a system service
	if checkAlgorandServiceCreated() {
		switch runtime.GOOS {
		case "linux":
			stopSystemdAlgorandService()
		case "darwin":
			stopLaunchdAlgorandService()
		default: // Unsupported OS
			fmt.Println("Unsupported OS.")
			os.Exit(1)
		}

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

func stopLaunchdAlgorandService() {
	fmt.Println("Stopping algod using launchd...")
	cmd := exec.Command("launchctl", "bootout", "system/com.algorand.algod")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Failed to stop algod service: %v\n", err)
		cobra.CheckErr(err)
	}
	fmt.Println("Algod service stopped.")
}

func stopSystemdAlgorandService() {
	fmt.Println("Stopping algod using systemctl...")
	cmd := exec.Command("systemctl", "stop", "algorand")
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Failed to stop algod service: %v\n", err)
		cobra.CheckErr(err)
	}
	fmt.Println("Algod service stopped.")
}
