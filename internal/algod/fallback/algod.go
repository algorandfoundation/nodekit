package fallback

import (
	"errors"
	"fmt"
	"github.com/algorandfoundation/algorun-tui/internal/algod/utils"
	"github.com/algorandfoundation/algorun-tui/internal/system"
	"os"
	"os/exec"
	"syscall"
)

func isRunning() (bool, error) {
	return false, errors.New("not implemented")
}
func Install() error {
	return system.RunAll(system.CmdsList{
		{"mkdir", "~/node"},
		{"sh", "-c", "cd ~/node"},
		{"wget", "https://raw.githubusercontent.com/algorand/go-algorand/rel/stable/cmd/updater/update.sh"},
		{"chmod", "744", "update.sh"},
		{"sh", "-c", "./update.sh -i -c stable -p ~/node -d ~/node/data -n"},
	})
}

func Start() error {
	// Algod is not available as a systemd service, start it directly
	fmt.Println("Starting algod directly...")

	// Check if ALGORAND_DATA environment variable is set
	fmt.Println("Checking if ALGORAND_DATA env var is set...")
	algorandData := os.Getenv("ALGORAND_DATA")

	if !utils.IsDataDir(algorandData) {
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
		return fmt.Errorf("Failed to start algod: %v\n", err)
	}
	return nil
}

func Stop() error {
	// Algod is not available as a systemd service, stop it directly
	fmt.Println("Stopping algod directly...")
	// Find the process ID of algod
	pid, err := findAlgodPID()
	if err != nil {
		return fmt.Errorf("Failed to find algod process: %v\n", err)
	}

	// Send SIGTERM to the process
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("Failed to find process with PID %d: %v\n", pid, err)
	}

	err = process.Signal(syscall.SIGTERM)
	if err != nil {
		return fmt.Errorf("Failed to send SIGTERM to process with PID %d: %v\n", pid, err)
	}

	fmt.Println("Sent SIGTERM to algod process.")
	return nil
}

func findAlgodPID() (int, error) {
	cmd := exec.Command("pgrep", "algod")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	var pid int
	_, err = fmt.Sscanf(string(output), "%d", &pid)
	if err != nil {
		return 0, fmt.Errorf("failed to parse PID: %v", err)
	}

	return pid, nil
}
