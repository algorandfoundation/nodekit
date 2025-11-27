package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/algorandfoundation/nodekit/internal/algod/config"
	"github.com/algorandfoundation/nodekit/internal/algod/telemetry"
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

// IsDataDir determines if the specified path is a valid Algorand
// data directory containing the "genesis.json" file.
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

	genesisFile := filepath.Join(path, "genesis.json")
	_, err = os.Stat(genesisFile)
	return err == nil
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

	file, err := os.ReadFile(filepath.Join(path, "algod.admin.token"))
	if err != nil {
		return token, err
	}

	token = strings.Replace(string(file), "\n", "", -1)
	return token, nil
}

func GetNetworkFromDataDir(path string) (string, error) {
	var network string
	file, err := os.ReadFile(filepath.Join(path, "genesis.json"))
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
	file, err := os.ReadFile(filepath.Join(path, "algod.pid"))
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
	file, err := os.ReadFile(filepath.Join(path, "algod.net"))
	if err != nil {
		return AlgodNetEndpointFileMissingAddress, nil
	}

	endpoint = "http://" + ReplaceEndpointUrl(string(file))

	return endpoint, nil
}

// GetLogConfigFromDataDir reads a logging configuration file from the
// specified data directory and unmarshals it into a telemetry.Config.
func GetLogConfigFromDataDir(path string) (*telemetry.Config, error) {
	var logConfig telemetry.Config
	file, err := os.ReadFile(filepath.Join(path, "logging.config"))
	if err != nil {
		return &logConfig, err
	}
	err = json.Unmarshal(file, &logConfig)
	if err != nil {
		return &logConfig, err
	}
	return &logConfig, nil
}

// WriteLogConfigToDataDir writes the provided telemetry log configuration to a file in the specified data directory.
// The configuration is formatted as indented JSON and saved to a file named "logging.config".
func WriteLogConfigToDataDir(path string, logConfig *telemetry.Config) error {
	file, err := json.MarshalIndent(logConfig, "", " ")
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(path, "logging.config"), file, 0o644)
	if err != nil {
		return err
	}
	return nil
}

// GetConfigFromDataDir reads a node configuration file from the
// specific data directory and unmarshals it into a config.Config.
func GetConfigFromDataDir(path string) (*config.Config, error) {
	var algodConfig config.Config

	file, err := os.ReadFile(filepath.Join(path, "config.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return &algodConfig, nil
		}
		return &algodConfig, err
	}

	err = json.Unmarshal(file, &algodConfig)
	if err != nil {
		return &algodConfig, err
	}

	return &algodConfig, nil
}

// WriteConfigToDataDir writes the provided node configuration to a file in the specified data directory.
// The configuration is formatted as indented JSON and saved to a file named "config.json".
func WriteConfigToDataDir(path string, algodConfig *config.Config) error {
	// Read an existing config and unmarshal it into a map, or make a new map.
	var currentConfigMap map[string]json.RawMessage
	configFile := filepath.Join(path, "config.json")
	file, err := os.ReadFile(configFile)
	if err == nil {
		if err := json.Unmarshal(file, &currentConfigMap); err != nil {
			return err
		}
	} else if os.IsPermission(err) {
		return err
	} else if !os.IsNotExist(err) {
		return err
	} else {
		currentConfigMap = make(map[string]json.RawMessage)
	}

	// We only want user-defined values (non-nil), so omitempty removes
	// everything else when unmarshaling into newConfigMap.
	tempConfigMap, err := json.Marshal(algodConfig)
	if err != nil {
		return err
	}
	var newConfigMap map[string]json.RawMessage
	if err := json.Unmarshal(tempConfigMap, &newConfigMap); err != nil {
		return err
	}

	// Update currentConfigMap with user-defined values.
	for key, value := range newConfigMap {
		currentConfigMap[key] = value
	}

	// Marshal and save the new config
	newConfig, err := json.MarshalIndent(currentConfigMap, "", "\t")
	if err != nil {
		return err
	}
	err = os.WriteFile(filepath.Join(path, "config.json"), newConfig, 0o664)
	if err != nil {
		return err
	}

	// If we're sudo'ed set permissions/ownership on the config file
	if system.IsSudo() {
		return setFilePermissions(configFile)
	}

	return nil
}

// setFilePermissions matches a file's user and group
// to its parent directory, and sets perms to 0o664.
// Will only succeed if run as root (sudo)
func setFilePermissions(filePath string) error {
	// Get the parent directory info
	dirPath := filepath.Dir(filePath)
	dirInfo, err := os.Stat(dirPath)
	if err != nil {
		return err
	}

	// Extract ownership info from the directory.
	// This is specific to Unix-like systems (Linux and macOS)
	stat, ok := dirInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("unable to get ownership details on file: %s", dirPath)
	}

	// Apply ownership permissions to file
	uid := int(stat.Uid)
	gid := int(stat.Gid)
	if err := os.Chown(filePath, uid, gid); err != nil {
		return err
	}
	// Also make sure group can write in the future
	if err := os.Chmod(filePath, 0o664); err != nil {
		return err
	}

	return nil
}

// ReplaceEndpointUrl replaces newline characters and wildcard IP addresses in a URL with a specific local address.
func ReplaceEndpointUrl(s string) string {
	s = strings.Replace(s, "\n", "", -1)
	s = strings.Replace(s, "0.0.0.0", "127.0.0.1", 1)
	s = strings.Replace(s, "[::]", "127.0.0.1", 1)
	return s
}

// ShowHybridPopUp returns true if a field in the NodeKitSettings doesn't exist or is false
func ShowHybridPopUp() bool {
	settings, err := GetNodekitSettings()
	if err != nil {
		return false
	}

	return !settings.DismissedNotices.HybridAvailable
}

// DontShowHybridPopUp sets the "HybridAvailable" notice to true so they don't see it again.
func DontShowHybridPopUp() error {
	settings, err := GetNodekitSettings()
	if err != nil {
		return err
	}

	settings.DismissedNotices.HybridAvailable = true

	return WriteNodekitSettings(settings)
}
