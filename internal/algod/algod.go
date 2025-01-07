package algod

import (
	"fmt"
	"github.com/algorandfoundation/nodekit/internal/algod/linux"
	"github.com/algorandfoundation/nodekit/internal/algod/mac"
	"github.com/algorandfoundation/nodekit/internal/system"
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
func IsInstalled() bool {
	return system.CmdExists("algod")
}

// IsRunning checks if the algod is currently running on the host operating system.
// It returns true if the application is running, or false if it is not or if an error occurs.
// This function supports Linux and macOS platforms. It returns an error for unsupported operating systems.
func IsRunning() bool {
	switch runtime.GOOS {
	case "linux", "darwin":
		return system.IsCmdRunning("algod")

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
		return mac.Upgrade(false)
	default:
		return fmt.Errorf(UnsupportedOSError)
	}
}

// Uninstall removes the Algorand software from the system based
// on the host operating system using appropriate methods.
func Uninstall(force bool) error {
	switch runtime.GOOS {
	case "linux":
		return linux.Uninstall()
	case "darwin":
		return mac.Uninstall(force)
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
		return mac.Start(false)
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
		return mac.Stop(false)
	default:
		return fmt.Errorf(UnsupportedOSError)
	}
}
