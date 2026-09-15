package statewalker

import (
	"context"
	"fmt"

	"github.com/algorandfoundation/nodekit/api"
	"github.com/charmbracelet/log"
)

// Env carries the shared state scenario phases operate on. Phases may mutate
// it (e.g. the network phase fills in DataDir and Client for later phases).
type Env struct {
	// NetworkDir is the root directory holding the provisioned network.
	NetworkDir string

	// DataDir is the data directory of the node scenarios drive (Primary).
	DataDir string

	// Speed is the round-time profile the network was provisioned with.
	Speed SpeedProfile

	// Client talks to the node at DataDir; set once the network is up.
	Client api.ClientWithResponsesInterface

	// Keep prevents teardown on failure (and on success for scenarios that
	// normally tear down), leaving the network up for inspection.
	Keep bool
}

// Phase is a single step of a scenario: Run drives the state change and
// Verify asserts the node reached it (via WaitFor/participation checks).
type Phase struct {
	// Name identifies the phase in logs and failure messages.
	Name string

	// Run drives the state change; nil for verify-only phases.
	Run func(ctx context.Context, env *Env) error

	// Verify asserts the state was reached; nil when Run self-verifies.
	Verify func(ctx context.Context, env *Env) error
}

// Scenario is an ordered composition of phases sharing one Env.
type Scenario struct {
	// Name identifies the scenario in logs and failure messages.
	Name string

	// Phases are executed strictly in order; the first failure aborts.
	Phases []Phase

	// KeepOnSuccess leaves the network up after a successful run (e.g. the
	// demo scenario, whose whole point is the resulting state).
	KeepOnSuccess bool
}

// Execute runs the phases in order. On failure the remaining phases are
// skipped and teardown is invoked unless Env.Keep is set. On success teardown
// is invoked unless the scenario keeps its network (KeepOnSuccess or Env.Keep).
func (s Scenario) Execute(ctx context.Context, env *Env, teardown func() error) error {
	for i, phase := range s.Phases {
		log.Info("Phase starting", "scenario", s.Name, "phase", phase.Name, "step", fmt.Sprintf("%d/%d", i+1, len(s.Phases)))
		if phase.Run != nil {
			if err := phase.Run(ctx, env); err != nil {
				return s.fail(env, teardown, phase.Name, "run", err)
			}
		}
		if phase.Verify != nil {
			if err := phase.Verify(ctx, env); err != nil {
				return s.fail(env, teardown, phase.Name, "verify", err)
			}
		}
		log.Info("Phase complete", "scenario", s.Name, "phase", phase.Name)
	}
	if s.KeepOnSuccess || env.Keep {
		log.Info("Scenario complete; network kept up", "scenario", s.Name, "datadir", env.DataDir)
		return nil
	}
	log.Info("Scenario complete; tearing down", "scenario", s.Name)
	return teardown()
}

// fail wraps a phase error, tearing the network down unless Env.Keep is set.
func (s Scenario) fail(env *Env, teardown func() error, phase string, step string, err error) error {
	wrapped := fmt.Errorf("scenario %s failed at phase %s (%s): %w", s.Name, phase, step, err)
	if env.Keep {
		log.Error("Phase failed; keeping the network up for inspection (--keep)", "phase", phase, "datadir", env.DataDir, "err", err)
		return wrapped
	}
	log.Error("Phase failed; tearing down", "phase", phase, "err", err)
	if terr := teardown(); terr != nil {
		log.Error("Teardown failed", "err", terr)
	}
	return wrapped
}
