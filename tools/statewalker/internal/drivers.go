package statewalker

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WalletAccount is a single row of `goal account list` on a node.
type WalletAccount struct {
	// Address is the account address.
	Address string

	// Online reports whether the account is currently registered online.
	Online bool

	// MicroAlgos is the account balance in microAlgos.
	MicroAlgos uint64
}

// ListAccounts returns the accounts of the default wallet on the given node.
func ListAccounts(dataDir string) ([]WalletAccount, error) {
	out, err := runner("goal", "account", "list", "-d", dataDir)
	if err != nil {
		return nil, fmt.Errorf("goal account list failed: %w\n%s", err, out)
	}
	var accounts []WalletAccount
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 4 || !strings.HasPrefix(fields[0], "[") {
			continue
		}
		account := WalletAccount{
			Online:  fields[0] == "[online]",
			Address: fields[2],
		}
		fmt.Sscanf(fields[3], "%d", &account.MicroAlgos)
		accounts = append(accounts, account)
	}
	if len(accounts) == 0 {
		return nil, fmt.Errorf("no wallet accounts found on %s:\n%s", dataDir, out)
	}
	return accounts, nil
}

// AddPartKey generates and installs a participation key for the account with
// the given validity window.
func AddPartKey(dataDir string, address string, firstValid uint64, lastValid uint64) (string, error) {
	out, err := runner("goal", "account", "addpartkey",
		"-d", dataDir,
		"-a", address,
		"--roundFirstValid", fmt.Sprintf("%d", firstValid),
		"--roundLastValid", fmt.Sprintf("%d", lastValid),
	)
	if err != nil {
		return out, fmt.Errorf("goal account addpartkey failed: %w\n%s", err, out)
	}
	return out, nil
}

// SetOnlineStatus broadcasts a keyreg transaction taking the account online or offline.
func SetOnlineStatus(dataDir string, address string, online bool) (string, error) {
	out, err := runner("goal", "account", "changeonlinestatus",
		"-d", dataDir,
		"-a", address,
		fmt.Sprintf("--online=%t", online),
	)
	if err != nil {
		return out, fmt.Errorf("goal account changeonlinestatus failed: %w\n%s", err, out)
	}
	return out, nil
}

// PartKeyInfo returns the `goal account partkeyinfo` report for the node.
func PartKeyInfo(dataDir string) (string, error) {
	out, err := runner("goal", "account", "partkeyinfo", "-d", dataDir)
	if err != nil {
		return out, fmt.Errorf("goal account partkeyinfo failed: %w\n%s", err, out)
	}
	return out, nil
}

// SendPayment sends a payment transaction between two addresses of the
// default wallet, advancing the round on networks without real consensus.
func SendPayment(dataDir string, from string, to string, microAlgos uint64) (string, error) {
	out, err := runner("goal", "clerk", "send",
		"-d", dataDir,
		"-f", from,
		"-t", to,
		"-a", fmt.Sprintf("%d", microAlgos),
	)
	if err != nil {
		return out, fmt.Errorf("goal clerk send failed: %w\n%s", err, out)
	}
	return out, nil
}

// StartNode starts a single node. When a peer address is given, the node
// dials it for gossip, required when starting an individual node of a
// private network, since only `goal network start` wires up the relay.
func StartNode(dataDir string, peer string) error {
	args := []string{"node", "start", "-d", dataDir}
	if peer != "" {
		args = append(args, "-p", peer)
	}
	out, err := runner("goal", args...)
	if err != nil {
		return fmt.Errorf("goal node start failed: %w\n%s", err, out)
	}
	return nil
}

// StopNode stops a single node.
func StopNode(dataDir string) error {
	out, err := runner("goal", "node", "stop", "-d", dataDir)
	if err != nil {
		return fmt.Errorf("goal node stop failed: %w\n%s", err, out)
	}
	return nil
}

// RestartNode restarts a single node, re-dialing the given peer when set.
func RestartNode(dataDir string, peer string) error {
	if err := StopNode(dataDir); err != nil {
		return err
	}
	return StartNode(dataDir, peer)
}

// RelayAddress returns the gossip listen address of the private network's
// relay, which individual nodes must dial when (re)started on their own.
func RelayAddress(dir string) (string, error) {
	raw, err := os.ReadFile(filepath.Join(dir, RelayNodeName, "algod-listen.net"))
	if err != nil {
		return "", fmt.Errorf("failed to read relay listen address: %w", err)
	}
	address := strings.TrimSpace(string(raw))
	address = strings.TrimPrefix(address, "http://")
	address = strings.TrimPrefix(address, "https://")
	return address, nil
}

// Catchup starts (or with abort, cancels) catchpoint catchup on a node.
func Catchup(dataDir string, catchpoint string) (string, error) {
	out, err := runner("goal", "node", "catchup", catchpoint, "-d", dataDir)
	if err != nil {
		return out, fmt.Errorf("goal node catchup failed: %w\n%s", err, out)
	}
	return out, nil
}

// GenesisID reads the node's genesis.json and returns "<network>-<id>",
// which is also the name of the ledger directory inside the data dir.
func GenesisID(dataDir string) (string, error) {
	raw, err := os.ReadFile(filepath.Join(dataDir, "genesis.json"))
	if err != nil {
		return "", err
	}
	var genesis struct {
		Network string `json:"network"`
		ID      string `json:"id"`
	}
	if err := json.Unmarshal(raw, &genesis); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s", genesis.Network, genesis.ID), nil
}

// GenesisProto reads the node's genesis.json and returns the consensus
// protocol the network was bootstrapped with.
func GenesisProto(dataDir string) (string, error) {
	raw, err := os.ReadFile(filepath.Join(dataDir, "genesis.json"))
	if err != nil {
		return "", err
	}
	var genesis struct {
		Proto string `json:"proto"`
	}
	if err := json.Unmarshal(raw, &genesis); err != nil {
		return "", err
	}
	return genesis.Proto, nil
}

// WipeLedger deletes the node's ledger databases so the node resyncs from the
// network on next start, while preserving installed participation keys
// (partregistry and *.partkey files). The node must be stopped first.
func WipeLedger(dataDir string) error {
	genesisID, err := GenesisID(dataDir)
	if err != nil {
		return err
	}
	ledgerDir := filepath.Join(dataDir, genesisID)
	entries, err := os.ReadDir(ledgerDir)
	if err != nil {
		return fmt.Errorf("ledger directory %s not found: %w", ledgerDir, err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, "ledger.") ||
			strings.HasPrefix(name, "crash.") ||
			strings.HasPrefix(name, "stateproof.") ||
			name == "catchpoints" {
			if err := os.RemoveAll(filepath.Join(ledgerDir, name)); err != nil {
				return err
			}
		}
	}
	return nil
}

// Protocols returns the consensus protocol table of the local algod as raw
// JSON, preserving number precision.
func Protocols(dataDir string) (map[string]map[string]json.RawMessage, error) {
	out, err := runner("goal", "protocols", "-d", dataDir)
	if err != nil {
		return nil, fmt.Errorf("goal protocols failed: %w\n%s", err, out)
	}
	// goal may print warnings before the JSON document
	start := strings.Index(out, "{")
	if start < 0 {
		return nil, fmt.Errorf("no JSON found in goal protocols output:\n%s", out)
	}
	var protocols map[string]map[string]json.RawMessage
	if err := json.Unmarshal([]byte(out[start:]), &protocols); err != nil {
		return nil, fmt.Errorf("failed to parse goal protocols output: %w", err)
	}
	return protocols, nil
}

// LocalnetGoal runs a goal command inside the algokit localnet container.
func LocalnetGoal(args ...string) (string, error) {
	out, err := runner("algokit", append([]string{"goal", "--"}, args...)...)
	if err != nil {
		return out, fmt.Errorf("algokit goal failed: %w\n%s", err, out)
	}
	return out, nil
}

// WriteConsensus installs a consensus.json override into the node's data dir.
func WriteConsensus(dataDir string, protocols map[string]map[string]json.RawMessage) error {
	raw, err := json.MarshalIndent(protocols, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dataDir, "consensus.json"), raw, 0644)
}
