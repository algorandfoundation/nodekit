package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/algorandfoundation/nodekit/internal/algod"
	statewalker "github.com/algorandfoundation/nodekit/tools/statewalker/internal"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// walkTimeout bounds how long walk scenarios wait for each expected state.
var walkTimeout time.Duration

// walkCmd is the parent for state-walking scenario commands. Scenarios only
// ever stop or wipe the Primary node; Node2 (anchor) and Relay stay untouched
// so the chain keeps advancing throughout.
var walkCmd = &cobra.Command{
	Use:   "walk",
	Short: "Walk the Primary node into sync-related states",
}

var walkSyncingCmd = &cobra.Command{
	Use:   "syncing",
	Short: "Wipe the Primary ledger so nodekit observes SYNCING while it re-joins",
	RunE: func(cmd *cobra.Command, args []string) error {
		dataDir, err := requirePrivate()
		if err != nil {
			return err
		}
		return driveSyncing(context.Background(), dataDir, walkTimeout)
	},
}

// driveSyncing stops the Primary node, wipes its ledger and restarts it so it
// resyncs from the anchor, waiting until it is stable and advancing again.
func driveSyncing(ctx context.Context, dataDir string, timeout time.Duration) error {
	relay, err := statewalker.RelayAddress(privateDir())
	if err != nil {
		return err
	}
	log.Info("Stopping Primary", "dir", dataDir)
	if err := statewalker.StopNode(dataDir); err != nil {
		return err
	}
	log.Info("Wiping Primary ledger")
	if err := statewalker.WipeLedger(dataDir); err != nil {
		return err
	}
	log.Info("Starting Primary (it will resync from the anchor)", "relay", relay)
	if err := statewalker.StartNode(dataDir, relay); err != nil {
		return err
	}
	client, err := getClient(ctx, dataDir)
	if err != nil {
		return err
	}
	log.Info("Waiting for SYNCING to be observable")
	if err := statewalker.WaitFor(ctx, client, statewalker.ExpectedState{State: algod.SyncingState}, timeout); err != nil {
		log.Warn("SYNCING was not observed; the node may have resynced instantly", "err", err)
	} else {
		log.Info("Node reports SYNCING; attach nodekit now to observe it")
	}
	log.Info("Waiting for the node to become stable again")
	if err := statewalker.WaitFor(ctx, client, statewalker.ExpectedState{State: algod.StableState}, timeout); err != nil {
		return err
	}
	log.Info("Verifying rounds advance")
	return statewalker.WaitForRoundsToAdvance(ctx, client, 2, timeout)
}

var walkFastCatchupCmd = &cobra.Command{
	Use:   "fast-catchup",
	Short: "Wipe the Primary ledger and start catchpoint catchup so nodekit observes FAST-CATCHUP",
	RunE: func(cmd *cobra.Command, args []string) error {
		dataDir, err := requirePrivate()
		if err != nil {
			return err
		}
		ctx := context.Background()

		// The anchor is only read here; it is never stopped or wiped.
		anchorClient, err := getClient(ctx, anchorDataDir())
		if err != nil {
			return err
		}
		log.Info("Looking up the latest catchpoint on the anchor")
		catchpoint, err := statewalker.LastCatchpoint(ctx, anchorClient)
		if err != nil {
			return err
		}
		if catchpoint == "" {
			return fmt.Errorf("no catchpoint available yet; the first one appears after CatchpointLookback+CatchpointInterval (~420 rounds), let the chain advance and retry")
		}
		log.Info("Using catchpoint", "catchpoint", catchpoint)

		relay, err := statewalker.RelayAddress(privateDir())
		if err != nil {
			return err
		}
		log.Info("Stopping Primary", "dir", dataDir)
		if err := statewalker.StopNode(dataDir); err != nil {
			return err
		}
		log.Info("Wiping Primary ledger")
		if err := statewalker.WipeLedger(dataDir); err != nil {
			return err
		}
		log.Info("Starting Primary", "relay", relay)
		if err := statewalker.StartNode(dataDir, relay); err != nil {
			return err
		}
		client, err := getClient(ctx, dataDir)
		if err != nil {
			return err
		}
		log.Info("Starting catchpoint catchup on Primary")
		// The freshly restarted node may not have connected to the relay yet,
		// so retry while algod reports an empty peer pool
		for attempt := 1; ; attempt++ {
			_, err = statewalker.Catchup(dataDir, catchpoint)
			if err == nil {
				break
			}
			if attempt >= 10 || !strings.Contains(err.Error(), "no peer pools available") {
				return err
			}
			log.Info("Node has no peers yet; retrying catchup", "attempt", attempt)
			time.Sleep(3 * time.Second)
		}
		log.Info("Waiting for FAST-CATCHUP to be observable")
		if err := statewalker.WaitFor(ctx, client, statewalker.ExpectedState{State: algod.FastCatchupState}, walkTimeout); err != nil {
			log.Warn("FAST-CATCHUP was not observed; catchup may have completed instantly", "err", err)
		} else {
			log.Info("Node reports FAST-CATCHUP; attach nodekit now to observe it")
		}
		log.Info("Waiting for the node to become stable again")
		if err := statewalker.WaitFor(ctx, client, statewalker.ExpectedState{State: algod.StableState}, walkTimeout); err != nil {
			return err
		}
		log.Info("Verifying rounds advance")
		return statewalker.WaitForRoundsToAdvance(ctx, client, 2, walkTimeout)
	},
}

func init() {
	walkCmd.PersistentFlags().DurationVar(&walkTimeout, "timeout", 5*time.Minute, "how long to wait for each expected state")
	walkCmd.AddCommand(walkSyncingCmd)
	walkCmd.AddCommand(walkFastCatchupCmd)
	rootCmd.AddCommand(walkCmd)
}
