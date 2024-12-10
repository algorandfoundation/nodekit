package algod

import (
	"fmt"
	"github.com/algorandfoundation/algorun-tui/internal/algod/linux"
	"github.com/algorandfoundation/algorun-tui/internal/algod/mac"
	"github.com/algorandfoundation/algorun-tui/internal/system"
	"os/exec"
	"runtime"
)

const UnsupportedOSError = "unsupported operating system"

// IsInstalled checks if the Algod software is installed on the system
// by verifying its presence and service setup.
func IsInstalled() bool {
	//If algod is not in the path
	if !system.CmdExists("algod") {
		return false
	}
	// If the service is listed
	switch runtime.GOOS {
	case "linux":
		return linux.IsService()
	case "darwin":
		return mac.IsService()
	default:
		fmt.Println("Unsupported operating system.")
		return false
	}
}

// IsRunning checks if the algod is currently running on the host operating system.
// It returns true if the application is running, or false if it is not or if an error occurs.
// This function supports Linux and macOS platforms. It returns an error for unsupported operating systems.
func IsRunning() bool {
	switch runtime.GOOS {
	case "linux", "darwin":
		fmt.Println("Checking if algod is running...")
		err := exec.Command("pgrep", "algod").Run()
		return err == nil
	default:
		return false
	}
}

// IsService determines if the Algorand service is configured as
// a system service on the current operating system.
func IsService() bool {
	switch runtime.GOOS {
	case "linux":
		return linux.IsService()
	case "darwin":
		return mac.IsService()
	default:
		return false
	}
}

// SetNetwork configures the network to the specified setting
// or returns an error on unsupported operating systems.
func SetNetwork(network string) error {
	return fmt.Errorf(UnsupportedOSError)
}

// Install installs Algorand software based on the host OS
// and returns an error if the installation fails or is unsupported.
func Install() error {
	switch runtime.GOOS {
	case "linux":
		return linux.Install()
	case "darwin":
		return mac.Install()
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
func Uninstall() error {
	switch runtime.GOOS {
	case "linux":
		return linux.Uninstall()
	case "darwin":
		return mac.Uninstall()
	default:
		return fmt.Errorf(UnsupportedOSError)
	}
}

// UpdateService updates the service configuration for the
// Algorand daemon based on the OS and reloads the service.
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

// EnsureService ensures the `algod` service is configured and running
// as a service based on the OS;
// Returns an error for unsupported systems.
func EnsureService() error {
	switch runtime.GOOS {
	case "darwin":
		return mac.EnsureService()
	default:
		return fmt.Errorf(UnsupportedOSError)
	}
}

// Start attempts to initiate the Algod service based on the
// host operating system. Returns an error for unsupported OS.
func Start() error {
	switch runtime.GOOS {
	case "linux":
		return linux.Start()
	case "darwin":
		return mac.Start()
	default: // Unsupported OS
		return fmt.Errorf(UnsupportedOSError)
	}
}

// Stop shuts down the Algorand algod system process based on the current operating system.
// Returns an error if the operation fails or the operating system is unsupported.
func Stop() error {
	switch runtime.GOOS {
	case "linux":
		return linux.Stop()
	case "darwin":
		return mac.Stop()
	default:
		return fmt.Errorf(UnsupportedOSError)
	}
}
