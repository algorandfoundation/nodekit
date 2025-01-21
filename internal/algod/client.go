package algod

import (
	"context"
	"errors"
	"github.com/algorandfoundation/nodekit/api"
	"github.com/algorandfoundation/nodekit/internal/algod/utils"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const InvalidDataDirMsg = "invalid data directory"
const ClientTimeoutMsg = "the client has timed out"

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

func WaitForClient(ctx context.Context, dataDir string, interval time.Duration, timeout time.Duration) (*api.ClientWithResponses, error) {
	var client *api.ClientWithResponses
	var err error
	dataDir, err = GetDataDir(dataDir)
	if err != nil {
		return client, err
	}
	// Try to fetch the client before waiting
	client, err = GetClient(dataDir)
	if err == nil {
		var resp api.ResponseInterface
		resp, err = client.GetStatusWithResponse(ctx)
		if err == nil && resp.StatusCode() == 200 {
			return client, nil
		}
	}
	// Wait for client to respond
	for {
		select {
		case <-ctx.Done():
			return client, nil
		case <-time.After(interval):
			client, err = GetClient(dataDir)
			if err == nil {
				var resp api.ResponseInterface
				resp, err = client.GetStatusWithResponse(ctx)
				if err == nil && resp.StatusCode() == 200 {
					return client, nil
				}
			}
		case <-time.After(timeout):
			return client, errors.New(ClientTimeoutMsg)
		}

	}
}
