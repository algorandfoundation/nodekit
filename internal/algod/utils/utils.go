package utils

import (
	"github.com/algorandfoundation/algorun-tui/internal/system"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

func IsDataDir(path string) bool {
	info, err := os.Stat(path)

	// Check if the path exists
	if os.IsNotExist(err) {
		return false
	}

	// Check if the path is a directory
	if !info.IsDir() {
		return false
	}

	paths := system.FindPathToFile(path, "algod.token")
	if len(paths) == 1 {
		return true
	}
	return false
}

func GetKnownPaths() []string {
	// Hardcoded paths known to be common Algorand data directories
	binPaths := []string{
		"/opt/homebrew/bin/algod",
		"/opt/homebrew/bin/algod",
	}

	var paths []string

	for _, path := range binPaths {
		if IsDataDir(path) {
			paths = append(paths, path)
		}
	}

	return paths
}

// GetKnownDataPaths Does a lazy check for Algorand data directories, based off of known common paths
func GetKnownDataPaths() []string {
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	// Hardcoded paths known to be common Algorand data directories
	commonAlgorandDataDirPaths := []string{
		"/var/lib/algorand",
		filepath.Join(home, "node", "data"),
		filepath.Join(home, ".algorand"),
	}

	var paths []string

	for _, path := range commonAlgorandDataDirPaths {
		if IsDataDir(path) {
			paths = append(paths, path)
		}
	}

	return paths
}
