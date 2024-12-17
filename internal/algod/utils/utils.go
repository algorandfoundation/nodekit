package utils

import (
	"github.com/algorandfoundation/algorun-tui/internal/system"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"time"
)

// IsDataDir determines if the specified path is a valid Algorand data directory containing an "algod.token" file.
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

// GetExpiresTime calculates and returns the expiration time of a vote based on rounds and time duration information.
// If the lastRound and roundTime are not zero, it computes the expiration using round difference and duration.
// Returns nil if the expiration time cannot be determined.
func GetExpiresTime(t system.Time, lastRound int, roundTime time.Duration, voteLastValid int) *time.Time {
	now := t.Now()
	var expires time.Time
	if lastRound != 0 &&
		roundTime != 0 {
		roundDiff := max(0, voteLastValid-lastRound)
		distance := int(roundTime) * roundDiff
		expires = now.Add(time.Duration(distance))
		return &expires
	}
	return nil
}
