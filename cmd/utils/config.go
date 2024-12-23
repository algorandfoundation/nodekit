package utils

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
