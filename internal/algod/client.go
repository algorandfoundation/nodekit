package algod

import (
	"errors"
	"github.com/algorandfoundation/nodekit/api"
	"github.com/algorandfoundation/nodekit/internal/algod/utils"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
	"os"
	"path/filepath"
	"runtime"
)

const InvalidDataDirMsg = "invalid data directory"

func GetDataDir(dataDir string) (string, error) {
	envDataDir := os.Getenv("ALGORAND_DATA")

	var defaultDataDir string
	switch runtime.GOOS {
	case "darwin":
		defaultDataDir = filepath.Join(os.Getenv("HOME"), ".algorand")
	case "linux":
		defaultDataDir = "/var/lib/algorand"
	default:
		return "", errors.New(UnsupportedOSError)
	}

	var resolvedDir string

	if dataDir == "" {
		if envDataDir == "" {
			resolvedDir = defaultDataDir
		} else {
			resolvedDir = envDataDir
		}
	} else {
		resolvedDir = dataDir
	}

	return resolvedDir, nil
}

// GetClient initializes and returns a new API client configured with the provided endpoint and access token.
func GetClient(dataDir string) (*api.ClientWithResponses, error) {
	resolvedDir, err := GetDataDir(dataDir)
	if err != nil {
		return nil, err
	}
	config, err := utils.ToDataFolderConfig(resolvedDir)
	if err != nil {
		return nil, err
	}

	apiToken, err := securityprovider.NewSecurityProviderApiKey("header", "X-Algo-API-Token", config.Token)
	if err != nil {
		return nil, err
	}
	return api.NewClientWithResponses(config.Endpoint, api.WithRequestEditorFn(apiToken.Intercept))
}
