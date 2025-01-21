package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/algorandfoundation/nodekit/internal/system"
	"github.com/spf13/cobra"
)

const AlgodNetEndpointFileMissingAddress = "missing://endpoint"

type DataFolderConfig struct {
	Path      string `json:"path"`
	BytesFree string `json:"bytesFree"`
	Token     string `json:"token"`
	Endpoint  string `json:"endpoint"`
	Network   string `json:"network"`
	PID       int    `json:"PID"`
}

func ToDataFolderConfig(path string) (DataFolderConfig, error) {
	var dataFolderConfig DataFolderConfig
	var err error
	if !IsDataDir(path) {
		return dataFolderConfig, nil
	}
	dataFolderConfig.Path = path
	dataFolderConfig.Token, err = GetTokenFromDataDir(path)
	if err != nil {
		return dataFolderConfig, err
	}
	dataFolderConfig.Network, err = GetNetworkFromDataDir(path)
	if err != nil {
		return dataFolderConfig, err
	}

	dataFolderConfig.Endpoint, _ = GetEndpointFromDataDir(path)
	dataFolderConfig.PID, _ = GetPidFromDataDir(path)

	return dataFolderConfig, nil
}

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

func GetTokenFromDataDir(path string) (string, error) {
	var token string

	file, err := os.ReadFile(path + "/algod.admin.token")
	if err != nil {
		return token, err
	}

	token = strings.Replace(string(file), "\n", "", -1)
	return token, nil
}

func GetNetworkFromDataDir(path string) (string, error) {
	var network string
	file, err := os.ReadFile(path + "/genesis.json")
	if err != nil {
		return network, err
	}
	var result map[string]interface{}
	err = json.Unmarshal(file, &result)
	if err != nil {
		return "", err
	}

	network = fmt.Sprintf("%s-%s", result["network"].(string), result["id"].(string))

	return network, nil
}

func GetPidFromDataDir(path string) (int, error) {
	var pid int
	file, err := os.ReadFile(path + "/algod.pid")
	if err != nil {
		return pid, err
	}

	pid, err = strconv.Atoi(strings.Replace(string(file), "\n", "", -1))
	if err != nil {
		return pid, err
	}

	return pid, nil
}

func GetEndpointFromDataDir(path string) (string, error) {
	var endpoint string
	file, err := os.ReadFile(path + "/algod.net")
	if err != nil {
		return AlgodNetEndpointFileMissingAddress, nil
	}

	endpoint = "http://" + ReplaceEndpointUrl(string(file))

	return endpoint, nil
}

// ReplaceEndpointUrl replaces newline characters and wildcard IP addresses in a URL with a specific local address.
func ReplaceEndpointUrl(s string) string {
	s = strings.Replace(s, "\n", "", -1)
	s = strings.Replace(s, "0.0.0.0", "127.0.0.1", 1)
	s = strings.Replace(s, "[::]", "127.0.0.1", 1)
	return s
}
