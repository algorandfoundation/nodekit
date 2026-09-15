package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/algorandfoundation/nodekit/internal/algod/participation"
	statewalker "github.com/algorandfoundation/nodekit/tools/statewalker/internal"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// stageSpeed selects the round-time profile of the staged network.
var stageSpeed string

// keepNetwork leaves the staged network up even when a phase fails or the
// scenario normally tears down.
var keepNetwork bool

// stageTimeout bounds how long each staged phase waits for its expected state.
var stageTimeout time.Duration

// demoVoteRounds keeps the demo upgrade vote visible long enough for tape capture.
const demoVoteRounds = 300

// stageCmd composes the network and state drivers into named, reproducible
// scenarios used identically by developers, CI and tape capture.
var stageCmd = &cobra.Command{
	Use:   "stage <demo|e2e|journey|tape>",
	Short: "Compose drivers into named, reproducible scenarios",
	Long: `Stages a named scenario on a fresh private network, verifying every phase
through /v2/status and /v2/participation:

  demo     seeded showcase state for tape capture: a long-validity online key,
           a near-expiry key, a zombie key and an upgrade vote in progress;
           the network stays up and DATADIR=<dir> is printed for the tapes.
  e2e      the PR fast suite: network up, partkey seed/online/offline/zombie,
           syncing walk, upgrade vote and a liveness invariant; tears down on
           completion unless --keep is given.
  journey  the containerized end-user lifecycle (install, catchup, online,
           upgrades, partkey regeneration).
  tape     the full-process TUI recording: a tuinet network advanced past a
           catchpoint plus a vhs-equipped container that records the entire
           end-user flow into assets/tapes/tui.gif (defaults to --speed fast).

Any existing tuinet private network is recreated so runs are deterministic.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		// The tape's ~420-round catchpoint pre-advance and on-camera catchup
		// only fit the recording budget at the fast cadence.
		if name == "tape" && !cmd.Flags().Changed("speed") {
			stageSpeed = string(statewalker.SpeedFast)
		}
		speed, err := statewalker.ParseSpeed(stageSpeed)
		if err != nil {
			return err
		}
		var scenario statewalker.Scenario
		teardown := stageTeardown
		switch name {
		case "demo":
			scenario = demoScenario()
		case "e2e":
			scenario = e2eScenario()
		case "journey":
			scenario = journeyScenario()
			teardown = journeyTeardown
		case "tape":
			scenario = tapeScenario()
			teardown = journeyTeardown
		default:
			return fmt.Errorf("unknown scenario %q: use demo | e2e | journey | tape", name)
		}
		env := &statewalker.Env{
			NetworkDir: networkDir,
			Speed:      speed,
			Keep:       keepNetwork,
		}
		if err := scenario.Execute(context.Background(), env, teardown); err != nil {
			return err
		}
		if scenario.KeepOnSuccess || keepNetwork {
			fmt.Printf("DATADIR=%s\n", env.DataDir)
			if name == "journey" {
				fmt.Printf("CONTAINER=%s\n", statewalker.JourneyContainer)
			}
		}
		return nil
	},
}

// stageTeardown stops and deletes the staged private network so the next
// stage run starts from a clean, deterministic state.
func stageTeardown() error {
	if _, err := os.Stat(privateDir()); os.IsNotExist(err) {
		return nil
	}
	if err := statewalker.StopPrivate(privateDir()); err != nil {
		log.Warn("Failed to stop the private network; deleting anyway", "err", err)
	}
	if err := statewalker.DeletePrivate(privateDir()); err != nil {
		return err
	}
	_ = os.Remove(modeFile())
	_ = os.Remove(speedFile())
	return nil
}

// stageNetworkPhase recreates the private network with the staged speed
// profile and fills Env.DataDir and Env.Client for the following phases.
// relayPublic exposes the relay on all interfaces for the journey container.
func stageNetworkPhase(relayPublic bool) statewalker.Phase {
	return statewalker.Phase{
		Name: "network",
		Run: func(ctx context.Context, env *statewalker.Env) error {
			report, err := statewalker.Preflight(statewalker.ModePrivate)
			fmt.Print(report)
			if err != nil {
				return err
			}
			if _, err := os.Stat(privateDir()); err == nil {
				log.Info("Existing private network found; recreating it for a deterministic stage")
				if err := stageTeardown(); err != nil {
					return err
				}
			}
			dataDir, err := upPrivate(ctx, env.Speed, stageTimeout, relayPublic)
			if err != nil {
				return err
			}
			env.DataDir = dataDir
			client, err := getClient(ctx, dataDir)
			if err != nil {
				return err
			}
			env.Client = client
			return nil
		},
		Verify: func(ctx context.Context, env *statewalker.Env) error {
			return statewalker.WaitForRoundsToAdvance(ctx, env.Client, 2, stageTimeout)
		},
	}
}

// countKeys returns how many participation keys the node holds per address.
func countKeys(ctx context.Context, env *statewalker.Env) (map[string]int, error) {
	keys, _, err := participation.GetList(ctx, env.Client)
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int)
	for _, key := range keys {
		counts[key.Address]++
	}
	return counts, nil
}

// demoScenario seeds a deterministic, visually rich node state for tape
// capture: a long-validity online key, a near-expiry key, a zombie key and an
// upgrade vote in progress. The network stays up on success.
func demoScenario() statewalker.Scenario {
	var online, offline statewalker.WalletAccount
	return statewalker.Scenario{
		Name:          "demo",
		KeepOnSuccess: true,
		Phases: []statewalker.Phase{
			stageNetworkPhase(false),
			{
				Name: "partkeys",
				Run: func(ctx context.Context, env *statewalker.Env) error {
					var err error
					if online, err = resolvePrimaryAccount(env.DataDir, "", true); err != nil {
						return err
					}
					if offline, err = resolvePrimaryAccount(env.DataDir, "", false); err != nil {
						return err
					}
					roundTime, err := statewalker.MeasureRoundTime(ctx, env.Client)
					if err != nil {
						return err
					}
					status, _, err := statewalker.CurrentStatus(ctx, env.Client)
					if err != nil {
						return err
					}
					longRounds := statewalker.DurationToRounds(720*time.Hour, roundTime)
					nearRounds := statewalker.DurationToRounds(72*time.Hour, roundTime)

					log.Info("Seeding long-validity key and registering online", "address", online.Address)
					if _, err := statewalker.AddPartKey(env.DataDir, online.Address, status.LastRound, status.LastRound+longRounds); err != nil {
						return err
					}
					if _, err := statewalker.SetOnlineStatus(env.DataDir, online.Address, true); err != nil {
						return err
					}
					log.Info("Seeding near-expiry key", "address", online.Address)
					if _, err := statewalker.AddPartKey(env.DataDir, online.Address, status.LastRound, status.LastRound+nearRounds); err != nil {
						return err
					}
					log.Info("Seeding zombie key for the offline account", "address", offline.Address)
					_, err = statewalker.AddPartKey(env.DataDir, offline.Address, status.LastRound, status.LastRound+longRounds)
					return err
				},
				Verify: func(ctx context.Context, env *statewalker.Env) error {
					counts, err := countKeys(ctx, env)
					if err != nil {
						return err
					}
					if counts[online.Address] < 2 {
						return fmt.Errorf("expected at least 2 keys for the online account %s, found %d", online.Address, counts[online.Address])
					}
					if counts[offline.Address] < 1 {
						return fmt.Errorf("expected a zombie key for the offline account %s, found none", offline.Address)
					}
					return statewalker.WaitForRoundsToAdvance(ctx, env.Client, 2, stageTimeout)
				},
			},
			{
				Name: "upgrade-vote",
				Run: func(ctx context.Context, env *statewalker.Env) error {
					return driveUpgradeVote(ctx, env.DataDir, env.Speed, demoVoteRounds, demoVoteRounds*8/10, demoVoteRounds, stageTimeout)
				},
				Verify: func(ctx context.Context, env *statewalker.Env) error {
					// The network restarted; reconnect before asserting
					client, err := getClient(ctx, env.DataDir)
					if err != nil {
						return err
					}
					env.Client = client
					return statewalker.WaitFor(ctx, env.Client, statewalker.ExpectedState{UpgradeVoting: true}, stageTimeout)
				},
			},
		},
	}
}

// tapeScenario provisions everything the full-process TUI recording needs: a
// tuinet network (genesis id tuinet-v1, so the QR registration modal renders)
// advanced past its first catchpoint, and a vhs-equipped journey container
// joined to it so the fresh container node can fast-catchup on camera. The
// recording and gif extraction phases land with internal/tape.go.
func tapeScenario() statewalker.Scenario {
	var relayPort string
	return statewalker.Scenario{
		Name: "tape",
		Phases: []statewalker.Phase{
			stageNetworkPhase(true),
			{
				Name: "catchpoint",
				Run: func(ctx context.Context, env *statewalker.Env) error {
					catchpoint, err := awaitCatchpoint(ctx)
					if err != nil {
						return err
					}
					log.Info("Catchpoint available for the on-camera fast catchup", "catchpoint", catchpoint)
					return nil
				},
				Verify: func(ctx context.Context, env *statewalker.Env) error {
					return statewalker.WaitForRoundsToAdvance(ctx, env.Client, 2, stageTimeout)
				},
			},
			{
				Name: "container",
				Run: func(ctx context.Context, env *statewalker.Env) error {
					var err error
					relayPort, err = provisionContainer()
					return err
				},
				Verify: func(ctx context.Context, env *statewalker.Env) error {
					if err := verifyRelayReachable(relayPort); err != nil {
						return err
					}
					// The recording toolchain must be present in the image
					if out, err := journeyExec("vhs --version && ttyd --version && ffmpeg -version | head -1"); err != nil {
						return fmt.Errorf("the recording toolchain is missing from the journey image: %w\n%s", err, out)
					}
					return nil
				},
			},
		},
	}
}

// e2eScenario is the PR fast suite: it exercises every state driver and the
// liveness invariant on a fresh network, then tears everything down.
func e2eScenario() statewalker.Scenario {
	var online, offline statewalker.WalletAccount
	return statewalker.Scenario{
		Name: "e2e",
		Phases: []statewalker.Phase{
			stageNetworkPhase(false),
			{
				Name: "partkeys-seed",
				Run: func(ctx context.Context, env *statewalker.Env) error {
					var err error
					if online, err = resolvePrimaryAccount(env.DataDir, "", true); err != nil {
						return err
					}
					if offline, err = resolvePrimaryAccount(env.DataDir, "", false); err != nil {
						return err
					}
					roundTime, err := statewalker.MeasureRoundTime(ctx, env.Client)
					if err != nil {
						return err
					}
					status, _, err := statewalker.CurrentStatus(ctx, env.Client)
					if err != nil {
						return err
					}
					for _, validity := range []time.Duration{72 * time.Hour, 720 * time.Hour} {
						rounds := statewalker.DurationToRounds(validity, roundTime)
						log.Info("Seeding key", "address", online.Address, "expiring", validity)
						if _, err := statewalker.AddPartKey(env.DataDir, online.Address, status.LastRound, status.LastRound+rounds); err != nil {
							return err
						}
					}
					return nil
				},
				Verify: func(ctx context.Context, env *statewalker.Env) error {
					counts, err := countKeys(ctx, env)
					if err != nil {
						return err
					}
					if counts[online.Address] < 2 {
						return fmt.Errorf("expected at least 2 seeded keys for %s, found %d", online.Address, counts[online.Address])
					}
					return nil
				},
			},
			{
				Name: "online",
				Run: func(ctx context.Context, env *statewalker.Env) error {
					log.Info("Registering account online", "address", online.Address)
					_, err := statewalker.SetOnlineStatus(env.DataDir, online.Address, true)
					return err
				},
				Verify: func(ctx context.Context, env *statewalker.Env) error {
					return statewalker.WaitForRoundsToAdvance(ctx, env.Client, 2, stageTimeout)
				},
			},
			{
				Name: "offline",
				Run: func(ctx context.Context, env *statewalker.Env) error {
					log.Info("Registering account offline", "address", online.Address)
					_, err := statewalker.SetOnlineStatus(env.DataDir, online.Address, false)
					return err
				},
				Verify: func(ctx context.Context, env *statewalker.Env) error {
					// The anchor stake must keep the chain alive on its own
					return statewalker.WaitForRoundsToAdvance(ctx, env.Client, 2, stageTimeout)
				},
			},
			{
				Name: "zombie",
				Run: func(ctx context.Context, env *statewalker.Env) error {
					status, _, err := statewalker.CurrentStatus(ctx, env.Client)
					if err != nil {
						return err
					}
					log.Info("Installing zombie key", "address", offline.Address)
					_, err = statewalker.AddPartKey(env.DataDir, offline.Address, status.LastRound, status.LastRound+1_000_000)
					return err
				},
				Verify: func(ctx context.Context, env *statewalker.Env) error {
					counts, err := countKeys(ctx, env)
					if err != nil {
						return err
					}
					if counts[offline.Address] < 1 {
						return fmt.Errorf("expected a zombie key for %s, found none", offline.Address)
					}
					return nil
				},
			},
			{
				Name: "syncing",
				Run: func(ctx context.Context, env *statewalker.Env) error {
					return driveSyncing(ctx, env.DataDir, stageTimeout)
				},
			},
			{
				Name: "upgrade-vote",
				Run: func(ctx context.Context, env *statewalker.Env) error {
					return driveUpgradeVote(ctx, env.DataDir, env.Speed, voteRounds, voteThreshold, upgradeDelay, stageTimeout)
				},
				Verify: func(ctx context.Context, env *statewalker.Env) error {
					// The network restarted; reconnect before asserting
					client, err := getClient(ctx, env.DataDir)
					if err != nil {
						return err
					}
					env.Client = client
					return statewalker.WaitFor(ctx, env.Client, statewalker.ExpectedState{UpgradeVoting: true}, stageTimeout)
				},
			},
			{
				Name: "liveness",
				Verify: func(ctx context.Context, env *statewalker.Env) error {
					return statewalker.WaitForRoundsToAdvance(ctx, env.Client, 5, stageTimeout)
				},
			},
		},
	}
}

func init() {
	stageCmd.Flags().StringVar(&stageSpeed, "speed", string(statewalker.SpeedReal), "round-time profile of the staged network: fast | real")
	stageCmd.Flags().BoolVar(&keepNetwork, "keep", false, "keep the network up even on failure or when the scenario normally tears down")
	stageCmd.Flags().DurationVar(&stageTimeout, "timeout", 5*time.Minute, "how long each phase waits for its expected state")
	rootCmd.AddCommand(stageCmd)
}
