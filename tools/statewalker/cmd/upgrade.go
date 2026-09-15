package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	statewalker "github.com/algorandfoundation/nodekit/tools/statewalker/internal"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// upgradeProtoID is the synthetic successor protocol declared in consensus.json.
const upgradeProtoID = "statewalker-upgrade-v1"

// voteRounds is the length of the upgrade voting window in rounds.
var voteRounds uint64

// voteThreshold is the number of yes votes required for the upgrade to pass.
var voteThreshold uint64

// upgradeDelay is the number of rounds between approval and activation.
var upgradeDelay uint64

// upgradeTimeout bounds how long `upgrade vote` waits for voting to appear.
var upgradeTimeout time.Duration

// upgradeCmd is the parent for consensus upgrade scenario commands.
var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Drive consensus upgrade voting on the private network",
}

// allNodeDirs returns the data dirs of every node in the private network.
func allNodeDirs() []string {
	return []string{
		statewalker.NodeDataDir(privateDir(), statewalker.RelayNodeName),
		statewalker.NodeDataDir(privateDir(), statewalker.PrimaryNodeName),
		statewalker.NodeDataDir(privateDir(), statewalker.SecondaryNodeName),
	}
}

var upgradeVoteCmd = &cobra.Command{
	Use:   "vote",
	Short: "Install a consensus.json declaring an approved upgrade and restart the network",
	Long: `Clones the network's genesis consensus protocol into a synthetic successor
and declares it in ApprovedUpgrades with a short voting window, then installs the
resulting consensus.json on every node and restarts the network. Participation
nodes start voting for the upgrade, so /v2/status (and the nodekit TUI) report
upgrade vote rounds and yes/no counts.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dataDir, err := requirePrivate()
		if err != nil {
			return err
		}
		ctx := context.Background()
		if err := driveUpgradeVote(ctx, dataDir, persistedSpeed(), voteRounds, voteThreshold, upgradeDelay, upgradeTimeout); err != nil {
			return err
		}
		fmt.Printf("\nAttach nodekit to observe the upgrade panel:\n\n  nodekit -d %s\n\n", dataDir)
		return nil
	},
}

// driveUpgradeVote installs the consensus override declaring an approved
// synthetic upgrade with the given voting window on every node, restarts the
// network and waits until /v2/status reports the vote in progress. The speed
// profile's overrides are re-merged into both protocol entries so accelerated
// round times survive the consensus rewrite and the upgrade itself.
func driveUpgradeVote(ctx context.Context, dataDir string, speed statewalker.SpeedProfile, voteRounds uint64, voteThreshold uint64, upgradeDelay uint64, timeout time.Duration) error {
	proto, err := statewalker.GenesisProto(dataDir)
	if err != nil {
		return err
	}
	log.Info("Genesis protocol", "proto", proto)
	protocols, err := statewalker.Protocols(dataDir)
	if err != nil {
		return err
	}
	current, ok := protocols[proto]
	if !ok {
		return fmt.Errorf("genesis protocol %s not found in the local algod protocol table", proto)
	}

	// Clone the current protocol as the synthetic successor
	successor := make(map[string]json.RawMessage, len(current))
	for key, value := range current {
		successor[key] = value
	}
	successor["ApprovedUpgrades"] = json.RawMessage("{}")

	// Shorten the voting window on the current protocol and approve the successor
	override := map[string]json.RawMessage{}
	for key, value := range current {
		override[key] = value
	}
	override["UpgradeVoteRounds"] = number(voteRounds)
	override["UpgradeThreshold"] = number(voteThreshold)
	override["MinUpgradeWaitRounds"] = number(upgradeDelay)
	override["DefaultUpgradeWaitRounds"] = number(upgradeDelay)
	approved, err := json.Marshal(map[string]uint64{upgradeProtoID: upgradeDelay})
	if err != nil {
		return err
	}
	override["ApprovedUpgrades"] = approved

	// Keep the accelerated round times across the rewrite and the upgrade
	for key, value := range speed.ConsensusOverrides() {
		override[key] = value
		successor[key] = value
	}

	consensus := map[string]map[string]json.RawMessage{
		proto:          override,
		upgradeProtoID: successor,
	}

	for _, nodeDir := range allNodeDirs() {
		log.Info("Installing consensus.json", "dir", nodeDir)
		if err := statewalker.WriteConsensus(nodeDir, consensus); err != nil {
			return err
		}
	}

	log.Info("Restarting the network to load the consensus override")
	if err := statewalker.StopPrivate(privateDir()); err != nil {
		return err
	}
	if err := statewalker.StartPrivate(privateDir()); err != nil {
		return err
	}

	client, err := getClient(ctx, dataDir)
	if err != nil {
		return err
	}
	log.Info("Waiting for upgrade voting to appear in /v2/status")
	if err := statewalker.WaitFor(ctx, client, statewalker.ExpectedState{UpgradeVoting: true}, timeout); err != nil {
		return err
	}
	status, _, err := statewalker.CurrentStatus(ctx, client)
	if err != nil {
		return err
	}
	log.Info("Upgrade voting in progress",
		"voteRounds", status.UpgradeVoteRounds,
		"yesVotes", status.UpgradeYesVotes,
		"noVotes", status.UpgradeNoVotes,
		"votesRequired", status.UpgradeVotesRequired,
	)
	return nil
}

var upgradeClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Remove the consensus.json overrides and restart the network",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := requirePrivate(); err != nil {
			return err
		}
		for _, nodeDir := range allNodeDirs() {
			path := filepath.Join(nodeDir, "consensus.json")
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return err
			}
			log.Info("Removed consensus override", "path", path)
		}
		log.Info("Restarting the network")
		if err := statewalker.StopPrivate(privateDir()); err != nil {
			return err
		}
		return statewalker.StartPrivate(privateDir())
	},
}

// number marshals a uint64 into a raw JSON number.
func number(value uint64) json.RawMessage {
	return json.RawMessage(fmt.Sprintf("%d", value))
}

func init() {
	upgradeVoteCmd.Flags().Uint64Var(&voteRounds, "vote-rounds", 100, "length of the voting window in rounds")
	upgradeVoteCmd.Flags().Uint64Var(&voteThreshold, "threshold", 80, "yes votes required for the upgrade to pass")
	upgradeVoteCmd.Flags().Uint64Var(&upgradeDelay, "delay", 100, "rounds between approval and activation")
	upgradeVoteCmd.Flags().DurationVar(&upgradeTimeout, "timeout", 3*time.Minute, "how long to wait for voting to appear")
	upgradeCmd.AddCommand(upgradeVoteCmd)
	upgradeCmd.AddCommand(upgradeClearCmd)
	rootCmd.AddCommand(upgradeCmd)
}
