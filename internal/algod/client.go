package algod

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/algorandfoundation/nodekit/api"
	"github.com/algorandfoundation/nodekit/internal/algod/utils"
	"github.com/charmbracelet/log"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
)

const InvalidDataDirMsg = "invalid data directory"
const ClientTimeoutMsg = "timed out while waiting for the node"

func GetDataDir(dataDir string) (string, error) {
	// Priority:
	// 1. Use provided `-d` directory
	// 2. Use environment variable `ALGORAND_DATA`
	// 3. Use default given by nodekit
	if dataDir == "" {
		dataDir = os.Getenv("ALGORAND_DATA")

		if dataDir == "" {
			switch runtime.GOOS {
			case "darwin":
				dataDir = filepath.Join(os.Getenv("HOME"), ".algorand")
			case "linux":
				dataDir = "/var/lib/algorand"
			default:
				return "", errors.New(UnsupportedOSError)
			}
		}
	}

	return dataDir, nil
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
	log.Info(fmt.Sprintf("Waiting for the node (up to %s)", timeout))
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
	timeoutTimer := time.After(timeout)
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
		case <-timeoutTimer:
			return client, errors.New(ClientTimeoutMsg)
		}
	}
}
