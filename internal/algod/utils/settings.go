package utils

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/algorandfoundation/nodekit/internal/system"
)

const NodeKitSettingsJSONFile = ".nodekit.json"

type Settings struct {
	DismissedNotices struct {
		HybridAvailable bool `json:",omitempty"`
	}
}

func GetNodekitSettings() (Settings, error) {
	var settings Settings

	home, err := os.UserHomeDir()
	if err != nil {
		// Can't identify home directory
		return settings, err
	}

	settingsFile, err := os.ReadFile(filepath.Join(home, NodeKitSettingsJSONFile))
	if err != nil {
		if os.IsNotExist(err) {
			return settings, nil
		}
		return settings, err
	}

	err = json.Unmarshal(settingsFile, &settings)
	if err != nil {
		return settings, err
	}

	return settings, nil
}

func WriteNodekitSettings(settings Settings) error {
	home, err := os.UserHomeDir()
	if err != nil {
		// Can't identify home directory
		return err
	}

	settingsFile := filepath.Join(home, NodeKitSettingsJSONFile)

	// Marshal and save the new settings
	newSettings, err := json.MarshalIndent(settings, "", "\t")
	if err != nil {
		return err
	}
	err = os.WriteFile(settingsFile, newSettings, 0o664)
	if err != nil {
		return err
	}

	// If we're sudo'ed set permissions/ownership on the config file
	if system.IsSudo() {
		return setFilePermissions(settingsFile)
	}

	return nil
}
