package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/algorandfoundation/nodekit/api"
	"github.com/algorandfoundation/nodekit/internal/algod"
	statewalker "github.com/algorandfoundation/nodekit/tools/statewalker/internal"
)

// requirePrivate ensures a private network is provisioned and returns the
// Primary node's data directory. Scenario commands only ever target Primary;
// Node2 (the anchor) and Relay are reserved to keep the chain alive.
func requirePrivate() (string, error) {
	m, err := resolveMode()
	if err != nil {
		return "", err
	}
	if m != statewalker.ModePrivate {
		return "", fmt.Errorf("this command requires the private network (current mode: %s); run `statewalker network up --mode private`", m)
	}
	return statewalker.NodeDataDir(privateDir(), statewalker.PrimaryNodeName), nil
}

// anchorDataDir returns the data directory of the Node2 anchor, used
// read-only (e.g. to look up catchpoints); it is never stopped or wiped.
func anchorDataDir() string {
	return statewalker.NodeDataDir(privateDir(), statewalker.SecondaryNodeName)
}

// getClient connects to the node at dataDir, waiting briefly for it to answer.
func getClient(ctx context.Context, dataDir string) (*api.ClientWithResponses, error) {
	return algod.WaitForClient(ctx, dataDir, time.Second, 30*time.Second)
}

// resolvePrimaryAccount returns the Primary wallet account to target: the
// explicit address when given (validated against Primary's wallet), otherwise
// the first account matching the online expectation.
func resolvePrimaryAccount(dataDir string, address string, online bool) (statewalker.WalletAccount, error) {
	accounts, err := statewalker.ListAccounts(dataDir)
	if err != nil {
		return statewalker.WalletAccount{}, err
	}
	if address != "" {
		for _, account := range accounts {
			if account.Address == address {
				return account, nil
			}
		}
		return statewalker.WalletAccount{}, fmt.Errorf("address %s is not a Primary wallet account; scenarios must not touch anchor (Node2) stake", address)
	}
	for _, account := range accounts {
		if account.Online == online {
			return account, nil
		}
	}
	return statewalker.WalletAccount{}, fmt.Errorf("no Primary wallet account found with online=%t", online)
}
