package algod

import (
	"fmt"
	"github.com/algorandfoundation/nodekit/internal/algod/fallback"
	"github.com/algorandfoundation/nodekit/internal/algod/linux"
	"github.com/algorandfoundation/nodekit/internal/algod/mac"
	"github.com/algorandfoundation/nodekit/internal/algod/utils"
	"github.com/algorandfoundation/nodekit/internal/system"
	"os"
	"runtime"
)

// UnsupportedOSError indicates that the current operating system is not supported for the requested operation.
const UnsupportedOSError = "unsupported operating system"

// InvalidStatusResponseError represents an error message indicating an invalid response status was encountered.
const InvalidStatusResponseError = "invalid status response"

// InvalidVersionResponseError represents an error message for an invalid response from the version endpoint.
const InvalidVersionResponseError = "invalid version response"

// IsInstalled checks if the Algod software is installed on the system
// by verifying its presence and service setup.
func IsInstalled(dataDir string) bool {
	return system.CmdExists("algod") && IsService(dataDir) && utils.IsDataDir(dataDir)
}

// IsRunning checks if the algod is currently running on the host operating system.
// It returns true if the application is running, or false if it is not or if an error occurs.
// This function supports Linux and macOS platforms. It returns an error for unsupported operating systems.
func IsRunning(dataDir string) bool {
	switch runtime.GOOS {
	case "linux", "darwin":
		// Simple check for algod process when nothing is passed in
		if dataDir == "" {
			return system.IsCmdRunning("algod")
		}

		// Try to discover PID from data directory
		pid, err := utils.GetPidFromDataDir(dataDir)

		// If the file doesn't exist or the pid is 0, it is not running
		if os.IsNotExist(err) || pid == 0 {
			return false
		}

		// Try to find the PID
		_, err = os.FindProcess(pid)
		if err == nil {
			// Successfully found the data directory pid
			return true
		}

		return false
	default:
		return false
	}
}

// IsService determines if the Algorand service is configured as
// a system service on the current operating system.
func IsService(dataDir string) bool {
	switch runtime.GOOS {
	case "linux":
		return linux.IsService(dataDir)
	case "darwin":
		return mac.IsService(dataDir)
	default:
		return false
	}
}

// SetNetwork configures the network to the specified setting
func SetNetwork(dataDir string, network string) error {
	return fallback.SetNetwork(dataDir, network, false)
}

// Install installs Algorand software based on the host OS
// and returns an error if the installation fails or is unsupported.
func Install(dataDir string, network string, force bool) error {
	switch runtime.GOOS {
	case "linux":
		return linux.Install(dataDir, network, force)
	case "darwin":
		return mac.Install(dataDir, network, force)
	default:
		return fmt.Errorf(UnsupportedOSError)
	}
}

// Update checks the operating system and performs an
// upgrade using OS-specific package managers, if supported.
func Update() error {
	switch runtime.GOOS {
	case "linux":
		return linux.Upgrade()
	case "darwin":
		return mac.Upgrade()
	default:
		return fmt.Errorf(UnsupportedOSError)
	}
}

// Uninstall removes the Algorand software from the system based
// on the host operating system using appropriate methods.
func Uninstall(dataDir string, force bool) error {
	switch runtime.GOOS {
	case "linux":
		return linux.Uninstall(dataDir, force)
	case "darwin":
		return mac.Uninstall(dataDir, force)
	default:
		return fmt.Errorf(UnsupportedOSError)
	}
}

// UpdateService updates the service configuration for the
// Algorand daemon based on the OS and reloads the service.
// WARNING: this requires elevation for system settings
func UpdateService(dataDirectoryPath string) error {
	switch runtime.GOOS {
	case "linux":
		return linux.UpdateService(dataDirectoryPath)
	case "darwin":
		return mac.UpdateService(dataDirectoryPath)
	default:
		return fmt.Errorf(UnsupportedOSError)
	}
}

// Start attempts to initiate the Algod service based on the
// host operating system. Returns an error for unsupported OS.
func Start(dataDir string) error {
	switch runtime.GOOS {
	case "linux":
		return linux.Start(dataDir)
	case "darwin":
		return mac.Start(dataDir)
	default: // Unsupported OS
		return fmt.Errorf(UnsupportedOSError)
	}
}

// Stop shuts down the Algorand algod system process based on the current operating system.
// Returns an error if the operation fails or the operating system is unsupported.
func Stop(dataDir string) error {
	switch runtime.GOOS {
	case "linux":
		return linux.Stop(dataDir)
	case "darwin":
		return mac.Stop(dataDir)
	default:
		return fmt.Errorf(UnsupportedOSError)
	}
}

func Pids(dataDir string) ([]int, error) {
	// Get all pids
	if dataDir == "" {
		return system.GetCmdPids("algod"), nil
	}
	// get the specific data directory pid
	pid, err := utils.GetPidFromDataDir(dataDir)
	return []int{pid}, err
}
