// Package statewalker contains the harness internals used by the statewalker CLI.
// It wraps the goal/algokit binaries to provision test networks and exposes
// helpers to verify algod state through the generated nodekit api client.
package statewalker

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Mode represents the network provisioning backend used by statewalker.
type Mode string

const (
	// ModePrivate provisions a private network via `goal network create`.
	ModePrivate Mode = "private"

	// ModeLocalnet provisions a dockerized network via `algokit localnet`.
	ModeLocalnet Mode = "localnet"
)

const (
	// PrivateNetworkName is the network name passed to `goal network create`.
	// It is deliberately "tuinet": goal assigns the genesis id "v1", so the
	// resulting genesis id is "tuinet-v1", which is already whitelisted by
	// nodekit's IsQREnabled() and mapped to localnet Lora URLs. The QR
	// registration modal therefore renders against this network with zero
	// production code changes.
	PrivateNetworkName = "tuinet"

	// PrimaryNodeName is the participation node nodekit should attach to.
	PrimaryNodeName = "Primary"

	// SecondaryNodeName is the second participation node, used by walk scenarios.
	SecondaryNodeName = "Node2"

	// RelayNodeName is the relay node of the private network.
	RelayNodeName = "Relay"

	// LocalnetEndpoint is the algod endpoint exposed by algokit localnet.
	LocalnetEndpoint = "127.0.0.1:4001"

	// LocalnetToken is the well-known algod token used by algokit localnet.
	LocalnetToken = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
)

// runner executes external commands; a variable so tests can stub it out.
var runner = func(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// lookPath resolves binaries on the PATH; a variable so tests can stub it out.
var lookPath = exec.LookPath

// Preflight verifies the binaries required by the given mode are installed
// and returns a human-readable report of the resolved versions.
func Preflight(mode Mode) (string, error) {
	var report strings.Builder
	var missing []string

	check := func(bin string, versionArgs ...string) {
		path, err := lookPath(bin)
		if err != nil {
			missing = append(missing, bin)
			return
		}
		out, err := runner(bin, versionArgs...)
		version := strings.TrimSpace(strings.SplitN(out, "\n", 2)[0])
		if err != nil || version == "" {
			version = "unknown version"
		}
		report.WriteString(fmt.Sprintf("%s: %s (%s)\n", bin, path, version))
	}

	switch mode {
	case ModePrivate:
		check("goal", "version")
		check("algod", "-v")
	case ModeLocalnet:
		check("algokit", "--version")
		check("docker", "--version")
	default:
		return "", fmt.Errorf("unknown mode: %s", mode)
	}

	if len(missing) > 0 {
		return report.String(), fmt.Errorf("missing required binaries for mode %q: %s", mode, strings.Join(missing, ", "))
	}
	return report.String(), nil
}

// CreatePrivate creates a private network rooted at dir from the given goal
// network template. The template is materialized next to the network dir so
// goal can consume it.
func CreatePrivate(dir string, template []byte) error {
	if _, err := os.Stat(dir); err == nil {
		return fmt.Errorf("network directory %s already exists; run `statewalker network down --delete` first", dir)
	}
	if err := os.MkdirAll(filepath.Dir(dir), 0755); err != nil {
		return err
	}
	templatePath := dir + ".template.json"
	if err := os.WriteFile(templatePath, template, 0644); err != nil {
		return err
	}
	out, err := runner("goal", "network", "create", "-n", PrivateNetworkName, "-r", dir, "-t", templatePath)
	if err != nil {
		return fmt.Errorf("goal network create failed: %w\n%s", err, out)
	}
	return nil
}

// StartPrivate starts all nodes of the private network rooted at dir.
func StartPrivate(dir string) error {
	out, err := runner("goal", "network", "start", "-r", dir)
	if err != nil {
		return fmt.Errorf("goal network start failed: %w\n%s", err, out)
	}
	return nil
}

// StopPrivate stops all nodes of the private network rooted at dir.
func StopPrivate(dir string) error {
	out, err := runner("goal", "network", "stop", "-r", dir)
	if err != nil {
		return fmt.Errorf("goal network stop failed: %w\n%s", err, out)
	}
	return nil
}

// DeletePrivate deletes the private network rooted at dir along with the
// materialized template file.
func DeletePrivate(dir string) error {
	out, err := runner("goal", "network", "delete", "-r", dir)
	if err != nil {
		return fmt.Errorf("goal network delete failed: %w\n%s", err, out)
	}
	_ = os.Remove(dir + ".template.json")
	return nil
}

// StatusPrivate returns the `goal network status` report for the network at dir.
func StatusPrivate(dir string) (string, error) {
	out, err := runner("goal", "network", "status", "-r", dir)
	if err != nil {
		return out, fmt.Errorf("goal network status failed: %w\n%s", err, out)
	}
	return out, nil
}

// NodeDataDir returns the data directory of a named node inside the network dir.
func NodeDataDir(dir string, node string) string {
	return filepath.Join(dir, node)
}

// MergeNodeConfig merges the given overrides into the node's config.json,
// creating the file when absent. Used to enable catchpoint tracking and other
// scenario-specific node settings.
func MergeNodeConfig(dataDir string, overrides map[string]interface{}) error {
	configPath := filepath.Join(dataDir, "config.json")
	config := make(map[string]interface{})
	if raw, err := os.ReadFile(configPath); err == nil {
		if err = json.Unmarshal(raw, &config); err != nil {
			return fmt.Errorf("failed to parse %s: %w", configPath, err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	for k, v := range overrides {
		config[k] = v
	}
	raw, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, raw, 0644)
}

// StartLocalnet starts the dockerized algokit localnet.
func StartLocalnet() error {
	out, err := runner("algokit", "localnet", "start")
	if err != nil {
		return fmt.Errorf("algokit localnet start failed: %w\n%s", err, out)
	}
	return nil
}

// StopLocalnet stops the dockerized algokit localnet.
func StopLocalnet() error {
	out, err := runner("algokit", "localnet", "stop")
	if err != nil {
		return fmt.Errorf("algokit localnet stop failed: %w\n%s", err, out)
	}
	return nil
}

// StatusLocalnet returns the `algokit localnet status` report.
func StatusLocalnet() (string, error) {
	out, err := runner("algokit", "localnet", "status")
	if err != nil {
		return out, fmt.Errorf("algokit localnet status failed: %w\n%s", err, out)
	}
	return out, nil
}

// WriteLocalnetShim writes a minimal data directory (algod.net plus token
// files) pointing at the localnet container so `nodekit -d <dir>` and the
// statewalker verifiers can attach to it like a regular node.
func WriteLocalnetShim(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	files := map[string]string{
		"algod.net":         LocalnetEndpoint,
		"algod.token":       LocalnetToken,
		"algod.admin.token": LocalnetToken,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}
