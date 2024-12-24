package algod

import (
	"github.com/algorandfoundation/algorun-tui/api"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/securityprovider"
)

// GetClient initializes and returns a new API client configured with the provided endpoint and access token.
func GetClient(endpoint string, token string) (*api.ClientWithResponses, error) {
	apiToken, err := securityprovider.NewSecurityProviderApiKey("header", "X-Algo-API-Token", token)
	if err != nil {
		return nil, err
	}
	return api.NewClientWithResponses(endpoint, api.WithRequestEditorFn(apiToken.Intercept))
}
