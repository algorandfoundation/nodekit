package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/algorandfoundation/nodekit/internal/algod"
	statewalker "github.com/algorandfoundation/nodekit/tools/statewalker/internal"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// mode selects the provisioning backend for network commands.
var mode string

// deleteNetwork removes the private network directory on `network down`.
var deleteNetwork bool

// upTimeout bounds how long `network up` waits for the node to become stable.
var upTimeout time.Duration

// netSpeed selects the round-time profile applied at private network creation.
var netSpeed string

// networkCmd is the parent for network lifecycle commands.
var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Provision and manage the test network",
}

// privateDir returns the goal network directory inside the network root.
func privateDir() string {
	return filepath.Join(networkDir, "private")
}

// localnetShimDir returns the shim data directory pointing at algokit localnet.
func localnetShimDir() string {
	return filepath.Join(networkDir, "localnet")
}

// modeFile persists the last provisioned mode so `down`/`status` do not need the flag.
func modeFile() string {
	return filepath.Join(networkDir, "mode")
}

// speedFile persists the speed profile the private network was created with.
func speedFile() string {
	return filepath.Join(networkDir, "speed")
}

// resolveMode returns the flag value or falls back to the persisted mode.
func resolveMode() (statewalker.Mode, error) {
	if mode != "" {
		return statewalker.Mode(mode), nil
	}
	raw, err := os.ReadFile(modeFile())
	if err != nil {
		return "", errors.New("no network mode given and no provisioned network found; use --mode private|localnet")
	}
	return statewalker.Mode(string(raw)), nil
}

// persistMode records the provisioned mode inside the network root.
func persistMode(m statewalker.Mode) error {
	if err := os.MkdirAll(networkDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(modeFile(), []byte(m), 0644)
}

// persistSpeed records the speed profile the private network was created with.
func persistSpeed(speed statewalker.SpeedProfile) error {
	if err := os.MkdirAll(networkDir, 0755); err != nil {
		return err
	}
	return os.WriteFile(speedFile(), []byte(speed), 0644)
}

// persistedSpeed returns the speed profile the private network was created
// with, defaulting to real when none was recorded.
func persistedSpeed() statewalker.SpeedProfile {
	raw, err := os.ReadFile(speedFile())
	if err != nil {
		return statewalker.SpeedReal
	}
	speed, err := statewalker.ParseSpeed(strings.TrimSpace(string(raw)))
	if err != nil {
		return statewalker.SpeedReal
	}
	return speed
}

// upPrivate provisions (creating it when absent) and starts the private
// network with the given speed profile, then waits until the Primary node
// reports a stable, advancing state. When relayPublic is set, the relay
// listens on all interfaces so docker containers (the journey node) can join.
// Returns Primary's data directory.
func upPrivate(ctx context.Context, speed statewalker.SpeedProfile, timeout time.Duration, relayPublic bool) (string, error) {
	dir := privateDir()
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Info("Creating private network", "dir", dir, "speed", speed)
		if err := statewalker.CreatePrivate(dir, PrivateTemplate); err != nil {
			return "", err
		}
		// Track catchpoints aggressively so fast-catchup scenarios are possible
		for _, node := range []string{statewalker.PrimaryNodeName, statewalker.SecondaryNodeName} {
			overrides := map[string]interface{}{
				"CatchpointTracking": 2,
				"CatchpointInterval": 100,
			}
			if err := statewalker.MergeNodeConfig(statewalker.NodeDataDir(dir, node), overrides); err != nil {
				return "", err
			}
		}
		// The relay must archive blocks and serve the ledger/block services,
		// otherwise catchpoint catchup has no peer pool to download from
		relayOverrides := map[string]interface{}{
			"Archival":            true,
			"EnableLedgerService": true,
			"EnableBlockService":  true,
			"CatchpointTracking":  2,
			"CatchpointInterval":  100,
		}
		if relayPublic {
			relayOverrides["NetAddress"] = "0.0.0.0:0"
		}
		if err := statewalker.MergeNodeConfig(statewalker.NodeDataDir(dir, statewalker.RelayNodeName), relayOverrides); err != nil {
			return "", err
		}
		if err := statewalker.ApplySpeed(speed, allNodeDirs()); err != nil {
			return "", err
		}
		if err := persistSpeed(speed); err != nil {
			return "", err
		}
	} else {
		log.Info("Reusing existing private network (the speed profile only applies at creation)", "dir", dir, "speed", persistedSpeed())
	}
	log.Info("Starting private network")
	if err := statewalker.StartPrivate(dir); err != nil {
		return "", err
	}
	dataDir := statewalker.NodeDataDir(dir, statewalker.PrimaryNodeName)
	if err := persistMode(statewalker.ModePrivate); err != nil {
		return "", err
	}
	client, err := algod.WaitForClient(ctx, dataDir, time.Second, timeout)
	if err != nil {
		return "", err
	}
	log.Info("Waiting for the node to report a stable state")
	if err := statewalker.WaitFor(ctx, client, statewalker.ExpectedState{State: algod.StableState}, timeout); err != nil {
		return "", err
	}
	log.Info("Verifying rounds advance without traffic")
	if err := statewalker.WaitForRoundsToAdvance(ctx, client, 2, timeout); err != nil {
		return "", err
	}
	return dataDir, nil
}

var networkUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Provision and start the test network",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := resolveMode()
		if err != nil {
			return err
		}
		report, err := statewalker.Preflight(m)
		fmt.Print(report)
		if err != nil {
			return err
		}

		ctx := context.Background()
		var dataDir string

		switch m {
		case statewalker.ModePrivate:
			speed, err := statewalker.ParseSpeed(netSpeed)
			if err != nil {
				return err
			}
			dataDir, err = upPrivate(ctx, speed, upTimeout, false)
			if err != nil {
				return err
			}
		case statewalker.ModeLocalnet:
			log.Info("Starting algokit localnet")
			if err := statewalker.StartLocalnet(); err != nil {
				return err
			}
			dataDir = localnetShimDir()
			if err := statewalker.WriteLocalnetShim(dataDir); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown mode: %s", m)
		}

		if m == statewalker.ModeLocalnet {
			if err := persistMode(m); err != nil {
				return err
			}
			client, err := algod.WaitForClient(ctx, dataDir, time.Second, upTimeout)
			if err != nil {
				return err
			}
			log.Info("Waiting for the node to report a stable state")
			if err := statewalker.WaitFor(ctx, client, statewalker.ExpectedState{State: algod.StableState}, upTimeout); err != nil {
				return err
			}
			log.Warn("localnet only advances rounds with traffic; run `statewalker traffic start` to keep the TUI moving")
		}

		log.Info("Network is up", "mode", m, "datadir", dataDir)
		fmt.Printf("\nAttach nodekit with:\n\n  nodekit -d %s\n\n", dataDir)
		return nil
	},
}

var networkDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Stop (and optionally delete) the test network",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := resolveMode()
		if err != nil {
			return err
		}
		switch m {
		case statewalker.ModePrivate:
			log.Info("Stopping private network", "dir", privateDir())
			if err := statewalker.StopPrivate(privateDir()); err != nil {
				return err
			}
			if deleteNetwork {
				log.Info("Deleting private network", "dir", privateDir())
				if err := statewalker.DeletePrivate(privateDir()); err != nil {
					return err
				}
				_ = os.Remove(modeFile())
				_ = os.Remove(speedFile())
			}
		case statewalker.ModeLocalnet:
			log.Info("Stopping algokit localnet")
			if err := statewalker.StopLocalnet(); err != nil {
				return err
			}
			if deleteNetwork {
				_ = os.RemoveAll(localnetShimDir())
				_ = os.Remove(modeFile())
			}
		default:
			return fmt.Errorf("unknown mode: %s", m)
		}
		log.Info("Network is down", "mode", m)
		return nil
	},
}

var networkStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the provisioning status of the test network",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := resolveMode()
		if err != nil {
			return err
		}
		report, err := statewalker.Preflight(m)
		fmt.Print(report)
		if err != nil {
			return err
		}
		switch m {
		case statewalker.ModePrivate:
			out, err := statewalker.StatusPrivate(privateDir())
			if err != nil {
				return err
			}
			fmt.Print(out)
			fmt.Printf("datadir: %s\n", statewalker.NodeDataDir(privateDir(), statewalker.PrimaryNodeName))
		case statewalker.ModeLocalnet:
			out, err := statewalker.StatusLocalnet()
			if err != nil {
				return err
			}
			fmt.Print(out)
			fmt.Printf("datadir: %s\n", localnetShimDir())
		default:
			return fmt.Errorf("unknown mode: %s", m)
		}
		return nil
	},
}

func init() {
	networkCmd.PersistentFlags().StringVar(&mode, "mode", "", "network backend: private | localnet (defaults to the provisioned one)")
	networkUpCmd.Flags().DurationVar(&upTimeout, "timeout", 2*time.Minute, "how long to wait for the network to stabilize")
	networkUpCmd.Flags().StringVar(&netSpeed, "speed", string(statewalker.SpeedReal), "round-time profile applied at private network creation: fast | real")
	networkDownCmd.Flags().BoolVar(&deleteNetwork, "delete", false, "delete the network after stopping it")
	networkCmd.AddCommand(networkUpCmd)
	networkCmd.AddCommand(networkDownCmd)
	networkCmd.AddCommand(networkStatusCmd)
}
