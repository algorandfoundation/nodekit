package algod

import (
	"errors"
	"github.com/algorandfoundation/algorun-tui/api"
	"github.com/algorandfoundation/algorun-tui/internal/algod/utils"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
	"os"
)

const InvalidDataDirMsg = "invalid data directory"

// GetClient initializes and returns a new API client configured with the provided endpoint and access token.
func GetClient(dataDir string) (*api.ClientWithResponses, error) {
	envDataDir := os.Getenv("ALGORAND_DATA")
	if envDataDir == "" && dataDir == "" {
		return nil, errors.New(InvalidDataDirMsg)
	}

	var resolvedDir string
	if dataDir == "" {
		resolvedDir = envDataDir
	} else {
		resolvedDir = dataDir
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
