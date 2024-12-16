package utils

import (
	"encoding/json"
	"github.com/spf13/viper"
	"io"
	"os"
	"strings"
)

func InitConfig() error {
	// Find home directory.
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	// Look for paths
	viper.AddConfigPath(".")
	viper.AddConfigPath(home)
	viper.AddConfigPath("/etc/algorun/")

	// Set Config Properties
	viper.SetConfigType("yaml")
	viper.SetConfigName(".algorun")
	viper.SetEnvPrefix("algorun")

	// Load Configurations
	viper.AutomaticEnv()
	_ = viper.ReadInConfig()

	// Check for algod
	loadedAlgod := viper.GetString("algod-endpoint")
	loadedToken := viper.GetString("algod-token")

	// Merge in the configuration from Algod
	conf, err := MergeAlgorandData(loadedAlgod, loadedToken)
	if err != nil {
		return err
	}

	if conf.Token != loadedToken {
		viper.Set("algod-token", conf.Token)
	}

	if conf.EndpointAddress != loadedAlgod {
		viper.Set("algod-endpoint", conf.EndpointAddress)
	}

	return nil
}

// Config represents the config.json file
type Config struct {
	EndpointAddress string `json:"EndpointAddress"`
}

// DaemonConfig represents the configuration settings for a daemon,
// including paths, network, token, and sub-configurations.
type DaemonConfig struct {
	DataDirectoryPath string `json:"data"`
	EndpointAddress   string `json:"endpoint"`
	Token             string `json:"token"`
}

func MergeAlgorandData(endpoint string, token string) (DaemonConfig, error) {
	result := DaemonConfig{
		DataDirectoryPath: "",
		EndpointAddress:   endpoint,
		Token:             token,
	}

	// Load ALGORAND_DATA/config.json
	algorandData, exists := os.LookupEnv("ALGORAND_DATA")

	// Load the Algorand Data Configuration
	if exists && algorandData != "" && endpoint == "" {
		// Placeholder for Struct
		var algodConfig Config

		dataConfigPath := algorandData + "/config.json"

		// Open the config.json File
		configFile, err := os.Open(dataConfigPath)
		if err != nil {
			return result, err
		}

		// Read the bytes of the File
		byteValue, _ := io.ReadAll(configFile)
		err = json.Unmarshal(byteValue, &algodConfig)
		if err != nil {
			return result, err
		}

		// Close the open handle
		err = configFile.Close()
		if err != nil {
			return result, err
		}

		// Check for endpoint address
		if hasWildcardEndpointUrl(algodConfig.EndpointAddress) {
			algodConfig.EndpointAddress = replaceEndpointUrl(algodConfig.EndpointAddress)
		} else if algodConfig.EndpointAddress == "" {
			// Assume it is not set, try to discover the port from the network file
			networkPath := algorandData + "/algod.net"
			networkFile, err := os.Open(networkPath)
			if err != nil {
				return result, err
			}

			byteValue, err = io.ReadAll(networkFile)
			if err != nil {
				return result, err
			}

			if hasWildcardEndpointUrl(string(byteValue)) {
				algodConfig.EndpointAddress = replaceEndpointUrl(string(byteValue))
			} else {
				algodConfig.EndpointAddress = string(byteValue)
			}

		}
		if strings.Contains(algodConfig.EndpointAddress, ":0") {
			algodConfig.EndpointAddress = strings.Replace(algodConfig.EndpointAddress, ":0", ":8080", 1)
		}
		if token == "" {
			// Handle Token Path
			tokenPath := algorandData + "/algod.admin.token"

			tokenFile, err := os.Open(tokenPath)
			if err != nil {
				return result, err
			}

			byteValue, err = io.ReadAll(tokenFile)
			if err != nil {
				return result, err
			}

			result.Token = strings.Replace(string(byteValue), "\n", "", 1)
		}

		// Set the algod configuration
		result.EndpointAddress = "http://" + strings.Replace(algodConfig.EndpointAddress, "\n", "", 1)
		result.DataDirectoryPath = algorandData
	}
	return result, nil
}

func replaceEndpointUrl(s string) string {
	s = strings.Replace(s, "\n", "", 1)
	s = strings.Replace(s, "0.0.0.0", "127.0.0.1", 1)
	s = strings.Replace(s, "[::]", "127.0.0.1", 1)
	return s
}
func hasWildcardEndpointUrl(s string) bool {
	return strings.Contains(s, "0.0.0.0") || strings.Contains(s, "::")
}
