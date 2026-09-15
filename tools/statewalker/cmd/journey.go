package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/algorandfoundation/nodekit/api"
	"github.com/algorandfoundation/nodekit/internal/algod"
	statewalker "github.com/algorandfoundation/nodekit/tools/statewalker/internal"
	"github.com/charmbracelet/log"
)

// JourneyDockerfile is the embedded journey container Dockerfile, injected by main.
var JourneyDockerfile []byte

// nodekitOldTag is the pinned older nodekit release the journey installs
// first, so the real `nodekit upgrade` path has something to upgrade.
var nodekitOldTag string

// algodOldVersion optionally pins an older algorand apt package version to
// exercise the algod upgrade path (empty installs the latest).
var algodOldVersion string

const (
	// journeyKeyLifeRounds is the validity window of the first user partkey;
	// short enough that the regen phase reaches near-expiry within the run.
	journeyKeyLifeRounds = 700

	// journeyRegenMargin is how many rounds before expiry the user regenerates.
	journeyRegenMargin = 300

	// journeyFundAmount funds the journey account with enough stake for keyregs.
	journeyFundAmount = 10_000_000_000_000

	// catchpointTimeout bounds the wait for the first catchpoint (~420 rounds).
	catchpointTimeout = 20 * time.Minute

	// regenTimeout bounds the wait for the short key to near its expiry.
	regenTimeout = 30 * time.Minute

	// observeTimeout bounds waits for transient states (SYNCING, FAST-CATCHUP)
	// that may legitimately be missed when the node moves through them faster
	// than the poll interval; missing one is a warning, not a failure.
	observeTimeout = 60 * time.Second

	// journeyKmdDir pins the kmd data dir inside the container so the wallet
	// bootstrap and every goal account command agree on where kmd lives
	// (Debian's goal otherwise resolves it inconsistently under sudo).
	journeyKmdDir = "/var/lib/algorand/kmd-statewalker"
)

// journeyGoal prefixes a goal invocation inside the container: privileged
// (the data dir is owned by the algorand user) with the pinned kmd dir.
func journeyGoal(args string) string {
	return "sudo -n env ALGORAND_KMD=" + journeyKmdDir + " goal " + args
}

// journeyAddressPattern matches an Algorand address in goal output, anchored
// to the account-creation success line.
var journeyAddressPattern = regexp.MustCompile(`Created new account with address ([A-Z2-7]{58})`)

// journeyDir returns the shim data directory pointing at the container node.
func journeyDir() string {
	return filepath.Join(networkDir, "journey")
}

// journeyBackend returns the environment the journey runs in. Docker/Debian
// is the only backend today; a macOS backend can be swapped in here later
// without touching the scenario phases.
func journeyBackend() statewalker.Backend {
	return statewalker.DockerBackend{
		Container:  statewalker.JourneyContainer,
		Image:      statewalker.JourneyImage,
		Dockerfile: JourneyDockerfile,
	}
}

// journeyTeardown dumps the container diagnostics, removes the container and
// the shim dir, then tears the private network down.
func journeyTeardown() error {
	backend := journeyBackend()
	if backend.Exists() {
		dumpJourneyDiagnostics()
		if err := backend.Remove(); err != nil {
			log.Warn("Failed to remove the journey container", "err", err)
		}
	}
	_ = os.RemoveAll(journeyDir())
	return stageTeardown()
}

// dumpJourneyDiagnostics prints the container's algod service log tail and the
// last /v2/status so scenario failures are debuggable from the CI log.
func dumpJourneyDiagnostics() {
	backend := journeyBackend()
	if out, err := backend.ExecRoot("journalctl -u algorand --no-pager -n 40 2>/dev/null || tail -40 /var/lib/algorand/node.log 2>/dev/null || true"); err == nil && strings.TrimSpace(out) != "" {
		log.Info("Journey container algod log tail:\n" + out)
	}
	if logs := backend.Logs(20); strings.TrimSpace(logs) != "" {
		log.Info("Journey container stdout/stderr tail:\n" + logs)
	}
	if client, err := algod.GetClient(journeyDir()); err == nil {
		if status, _, err := statewalker.CurrentStatus(context.Background(), client); err == nil {
			log.Info("Journey node last status", "state", status.State, "round", status.LastRound)
		}
	}
}

// journeyExec runs a user-flow script as the journey's end-user account
// (non-root, passwordless sudo, noninteractive frontends, hang watchdog).
func journeyExec(script string, env ...string) (string, error) {
	return journeyBackend().Exec(script, env...)
}

// journeyExecRoot runs orchestration scaffolding (network rewiring, ledger
// wipes) with full privileges; end-user phases must use journeyExec instead.
func journeyExecRoot(script string) (string, error) {
	return journeyBackend().ExecRoot(script)
}

// journeyClient connects to the container node through the shim data dir.
func journeyClient(ctx context.Context) (*api.ClientWithResponses, error) {
	return getClient(ctx, journeyDir())
}

// provisionContainer builds and starts the journey container, returning the
// host relay port the container node should gossip to.
func provisionContainer() (string, error) {
	if err := journeyBackend().Provision(); err != nil {
		return "", err
	}
	relay, err := statewalker.RelayAddress(privateDir())
	if err != nil {
		return "", err
	}
	return relay[strings.LastIndex(relay, ":")+1:], nil
}

// verifyRelayReachable preflights that the host relay is reachable from
// inside the container.
func verifyRelayReachable(relayPort string) error {
	if _, err := journeyExec(fmt.Sprintf("timeout 5 bash -c '</dev/tcp/host.docker.internal/%s'", relayPort)); err != nil {
		return fmt.Errorf("the host relay (port %s) is not reachable from the container; "+
			"check that docker's host-gateway works and no firewall blocks the docker bridge: %w", relayPort, err)
	}
	return nil
}

// awaitCatchpoint blocks until the anchor node emits a catchpoint. The first
// one appears after CatchpointLookback+CatchpointInterval (~420 rounds).
func awaitCatchpoint(ctx context.Context) (string, error) {
	anchorClient, err := getClient(ctx, anchorDataDir())
	if err != nil {
		return "", err
	}
	log.Info("Waiting for the first catchpoint on the host network (~420 rounds)")
	deadline := time.Now().Add(catchpointTimeout)
	for {
		catchpoint, err := statewalker.LastCatchpoint(ctx, anchorClient)
		if err != nil {
			return "", err
		}
		if catchpoint != "" {
			return catchpoint, nil
		}
		if time.Now().After(deadline) {
			return "", fmt.Errorf("no catchpoint appeared within %s", catchpointTimeout)
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
}

// journeyScenario walks the full end-user lifecycle in a throwaway container
// joined to the host private network: fresh install of a pinned-old nodekit,
// sync, fast catchup, go online, real nodekit+algod upgrade, and partkey
// regeneration near expiry.
func journeyScenario() statewalker.Scenario {
	var relayPort string
	var userAddress string
	var voteLastValid uint64

	dataDir := func() string { return statewalker.JourneyDataDir }

	return statewalker.Scenario{
		Name: "journey",
		Phases: []statewalker.Phase{
			stageNetworkPhase(true),
			{
				Name: "container",
				Run: func(ctx context.Context, env *statewalker.Env) error {
					var err error
					relayPort, err = provisionContainer()
					return err
				},
				Verify: func(ctx context.Context, env *statewalker.Env) error {
					// Preflight: the host relay must be reachable from the container
					return verifyRelayReachable(relayPort)
				},
			},
			{
				Name: "install",
				Run: func(ctx context.Context, env *statewalker.Env) error {
					log.Info("Running the documented install.sh entrypoint (latest nodekit)")
					if _, err := journeyExec(
						"(curl -fsSL https://nodekit.run/install.sh || curl -fsSL https://raw.githubusercontent.com/algorandfoundation/nodekit/main/install.sh) | bash",
						"NODEKIT_SKIP_BOOTSTRAP=1", "NODEKIT_FORCE_INSTALL=1"); err != nil {
						return fmt.Errorf("upstream fetch failed (install.sh): %w", err)
					}
					latest, err := journeyExec("./nodekit --version")
					if err != nil {
						return err
					}
					log.Info("install.sh delivered nodekit", "version", strings.TrimSpace(latest))

					// The binary stays in the user's home dir, exactly where
					// install.sh dropped it: self-upgrade writes its temp file
					// next to the executable, which only works user-writably.
					log.Info("Pinning the old nodekit release", "tag", nodekitOldTag)
					if _, err := journeyExec(fmt.Sprintf(
						"curl -fsSL -o ./nodekit https://github.com/algorandfoundation/nodekit/releases/download/%s/nodekit-amd64-linux && chmod +x ./nodekit",
						nodekitOldTag)); err != nil {
						return fmt.Errorf("upstream fetch failed (nodekit %s): %w", nodekitOldTag, err)
					}

					log.Info("Installing algod via `nodekit install` (apt path)")
					if out, err := journeyExec("./nodekit install"); err != nil {
						return fmt.Errorf("nodekit install failed: %w\n%s", err, out)
					}
					if algodOldVersion != "" {
						log.Info("Pinning old algod package", "version", algodOldVersion)
						if out, err := journeyExec(fmt.Sprintf(
							"sudo -n apt-get install -y --allow-downgrades algorand=%s && sudo -n systemctl restart algorand", algodOldVersion)); err != nil {
							return fmt.Errorf("failed to pin algorand=%s: %w\n%s", algodOldVersion, err, out)
						}
					}
					return nil
				},
				Verify: func(ctx context.Context, env *statewalker.Env) error {
					version, err := journeyExec("./nodekit --version && algod -v | head -2 && systemctl is-active algorand")
					if err != nil {
						return fmt.Errorf("algod is not running after install: %w", err)
					}
					log.Info("Installed versions:\n" + version)
					if !strings.Contains(version, strings.TrimPrefix(nodekitOldTag, "v")) {
						return fmt.Errorf("expected pinned nodekit %s, got: %s", nodekitOldTag, version)
					}
					return nil
				},
			},
			{
				Name: "connect",
				Run: func(ctx context.Context, env *statewalker.Env) error {
					backend := journeyBackend()
					log.Info("Rewiring the container node onto the private network")
					if _, err := journeyExecRoot("systemctl stop algorand && rm -rf " + dataDir() + "/mainnet-v1.0 " + dataDir() + "/genesis.json"); err != nil {
						return err
					}
					if err := backend.CopyTo(filepath.Join(env.DataDir, "genesis.json"), dataDir()+"/genesis.json"); err != nil {
						return err
					}
					// Match the host consensus overrides (upgrade vote and/or speed profile)
					if _, err := os.Stat(filepath.Join(env.DataDir, "consensus.json")); err == nil {
						if err := backend.CopyTo(filepath.Join(env.DataDir, "consensus.json"), dataDir()+"/consensus.json"); err != nil {
							return err
						}
					}
					if err := backend.WriteFile(dataDir()+"/config.json",
						`{"DNSBootstrapID":"","EndpointAddress":"0.0.0.0:`+statewalker.JourneyRestPort+`","GossipFanout":1}`); err != nil {
						return err
					}
					// Point the service at the host relay via a systemd drop-in
					dropIn := "[Service]\nExecStart=\nExecStart=/usr/bin/algod -d " + dataDir() + " -p host.docker.internal:" + relayPort + "\n"
					if _, err := journeyExecRoot("mkdir -p /etc/systemd/system/algorand.service.d"); err != nil {
						return err
					}
					if err := backend.WriteFile("/etc/systemd/system/algorand.service.d/10-statewalker.conf", dropIn); err != nil {
						return err
					}
					if _, err := journeyExecRoot("chown -R algorand:algorand " + dataDir() + " && systemctl daemon-reload && systemctl start algorand"); err != nil {
						return err
					}
					// Wait for the REST endpoint and expose it to the host via a shim data dir
					if _, err := journeyExecRoot("timeout 30 bash -c 'until [ -s " + dataDir() + "/algod.admin.token ]; do sleep 1; done'"); err != nil {
						return fmt.Errorf("algod did not write its tokens: %w", err)
					}
					token, err := backend.ReadFile(dataDir() + "/algod.token")
					if err != nil {
						return err
					}
					adminToken, err := backend.ReadFile(dataDir() + "/algod.admin.token")
					if err != nil {
						return err
					}
					endpoint, err := backend.Endpoint(statewalker.JourneyRestPort)
					if err != nil {
						return err
					}
					if err := statewalker.WriteNodeShim(journeyDir(), endpoint, strings.TrimSpace(token), strings.TrimSpace(adminToken)); err != nil {
						return err
					}
					// nodekit only treats directories containing a genesis.json
					// as data dirs, so mirror the network genesis into the shim
					genesis, err := os.ReadFile(filepath.Join(env.DataDir, "genesis.json"))
					if err != nil {
						return err
					}
					if err := os.WriteFile(filepath.Join(journeyDir(), "genesis.json"), genesis, 0644); err != nil {
						return err
					}
					log.Info("Container node joined the private network", "shim", journeyDir(), "endpoint", endpoint)
					return nil
				},
				Verify: func(ctx context.Context, env *statewalker.Env) error {
					client, err := journeyClient(ctx)
					if err != nil {
						return err
					}
					if err := statewalker.WaitFor(ctx, client, statewalker.ExpectedState{State: algod.SyncingState}, observeTimeout); err != nil {
						log.Warn("SYNCING was not observed; the node may have synced instantly", "err", err)
					} else {
						log.Info("Container node reports SYNCING against the private network")
					}
					// The node must be talking to our chain (rounds advancing)
					return statewalker.WaitForRoundsToAdvance(ctx, client, 2, stageTimeout)
				},
			},
			{
				Name: "fast-catchup",
				Run: func(ctx context.Context, env *statewalker.Env) error {
					catchpoint, err := awaitCatchpoint(ctx)
					if err != nil {
						return err
					}
					log.Info("Using catchpoint", "catchpoint", catchpoint)

					genesisID, err := statewalker.GenesisID(env.DataDir)
					if err != nil {
						return err
					}
					log.Info("Wiping the container ledger so catchup starts from behind")
					if _, err := journeyExecRoot("systemctl stop algorand && rm -rf " + dataDir() + "/" + genesisID + " && systemctl start algorand"); err != nil {
						return err
					}
					client, err := journeyClient(ctx)
					if err != nil {
						return err
					}
					log.Info("Starting catchpoint catchup on the container node")
					for attempt := 1; ; attempt++ {
						message, _, err := algod.StartCatchup(ctx, client, catchpoint, nil)
						if err == nil {
							log.Info("Catchup accepted", "message", strings.TrimSpace(message))
							break
						}
						if attempt >= 20 {
							return fmt.Errorf("catchup did not start: %w", err)
						}
						time.Sleep(3 * time.Second)
					}
					return nil
				},
				Verify: func(ctx context.Context, env *statewalker.Env) error {
					client, err := journeyClient(ctx)
					if err != nil {
						return err
					}
					if err := statewalker.WaitFor(ctx, client, statewalker.ExpectedState{State: algod.FastCatchupState}, observeTimeout); err != nil {
						log.Warn("FAST-CATCHUP was not observed; catchup may have completed instantly", "err", err)
					} else {
						log.Info("Container node reports FAST-CATCHUP")
					}
					if err := statewalker.WaitFor(ctx, client, statewalker.ExpectedState{State: algod.StableState}, stageTimeout); err != nil {
						return err
					}
					return statewalker.WaitForRoundsToAdvance(ctx, client, 2, stageTimeout)
				},
			},
			{
				Name: "go-online",
				Run: func(ctx context.Context, env *statewalker.Env) error {
					// A fresh node has no kmd wallet, and `goal wallet new`
					// demands a TTY for its password prompt; create the
					// blank-password default wallet through the kmd REST API
					// so account operations stay fully non-interactive.
					log.Info("Creating the default wallet via the kmd API")
					if out, err := journeyExec(
						"sudo -n mkdir -p " + journeyKmdDir + " && sudo -n chmod 700 " + journeyKmdDir +
							" && " + journeyGoal("kmd start -t 0 -d "+dataDir()) +
							" && sudo -n bash -c 'timeout 15 bash -c \"until [ -s " + journeyKmdDir + "/kmd.net ]; do sleep 1; done\";" +
							" curl -sf -H \"X-KMD-API-Token: $(cat " + journeyKmdDir + "/kmd.token)\"" +
							" -d \"{\\\"wallet_name\\\":\\\"unencrypted-default-wallet\\\",\\\"wallet_password\\\":\\\"\\\",\\\"wallet_driver_name\\\":\\\"sqlite\\\"}\"" +
							" \"http://$(cat " + journeyKmdDir + "/kmd.net)/v1/wallet\"'"); err != nil {
						return fmt.Errorf("wallet creation failed: %w\n%s", err, out)
					}
					out, err := journeyExec(journeyGoal("account new statewalker-user -d " + dataDir()))
					if err != nil {
						return err
					}
					// goal may interleave harmless log noise on stderr, so pick
					// the address out of the success line instead of splitting.
					if match := journeyAddressPattern.FindStringSubmatch(out); match != nil {
						userAddress = match[1]
					}
					if userAddress == "" {
						return fmt.Errorf("could not find the new account address in: %s", out)
					}
					log.Info("Created user account in the container", "address", userAddress)

					funder, err := resolvePrimaryAccount(env.DataDir, "", true)
					if err != nil {
						return err
					}
					log.Info("Funding the user account from the host wallet", "from", funder.Address)
					if _, err := statewalker.SendPayment(env.DataDir, funder.Address, userAddress, journeyFundAmount); err != nil {
						return err
					}
					client, err := journeyClient(ctx)
					if err != nil {
						return err
					}
					if err := waitForBalance(ctx, client, userAddress, journeyFundAmount, stageTimeout); err != nil {
						return err
					}
					status, _, err := statewalker.CurrentStatus(ctx, client)
					if err != nil {
						return err
					}
					voteLastValid = status.LastRound + journeyKeyLifeRounds
					log.Info("Generating a short-validity partkey and registering online", "lastValid", voteLastValid)
					if out, err := journeyExec(fmt.Sprintf(
						"%s && %s",
						journeyGoal(fmt.Sprintf("account addpartkey -a %s --roundFirstValid 0 --roundLastValid %d -d %s", userAddress, voteLastValid, dataDir())),
						journeyGoal(fmt.Sprintf("account changeonlinestatus -a %s --online=true -d %s", userAddress, dataDir())))); err != nil {
						return fmt.Errorf("going online failed: %w\n%s", err, out)
					}
					return nil
				},
				Verify: func(ctx context.Context, env *statewalker.Env) error {
					client, err := journeyClient(ctx)
					if err != nil {
						return err
					}
					return waitForOnline(ctx, client, userAddress, 0, stageTimeout)
				},
			},
			{
				Name: "upgrade",
				Run: func(ctx context.Context, env *statewalker.Env) error {
					log.Info("Running the real `nodekit upgrade` from the pinned old release")
					out, err := journeyExec("./nodekit upgrade")
					if err != nil {
						if strings.Contains(out, ".nodekit.bak") {
							// Known PR #197 re-exec failure mode: the self-upgrade
							// succeeded but the re-exec resolved the deleted backup.
							log.Warn("nodekit upgrade hit the known PR #197 re-exec bug; " +
								"the self-upgrade succeeded but the algod step was skipped. " +
								"Completing the algod upgrade via the same apt path nodekit uses.")
							if out, err := journeyExec("sudo -n apt-get update -q && sudo -n apt-get install -y --only-upgrade algorand && (systemctl is-active algorand || sudo -n systemctl start algorand)"); err != nil {
								return fmt.Errorf("algod upgrade failed: %w\n%s", err, out)
							}
						} else {
							return fmt.Errorf("nodekit upgrade failed: %w\n%s", err, out)
						}
					} else {
						log.Info("nodekit upgrade completed:\n" + out)
					}
					version, err := journeyExec("./nodekit --version")
					if err != nil {
						return err
					}
					if strings.Contains(version, strings.TrimPrefix(nodekitOldTag, "v")) {
						return fmt.Errorf("nodekit did not self-upgrade; still at %s", strings.TrimSpace(version))
					}
					log.Info("nodekit self-upgraded", "version", strings.TrimSpace(version))
					return nil
				},
				Verify: func(ctx context.Context, env *statewalker.Env) error {
					if _, err := journeyExec("systemctl is-active algorand"); err != nil {
						return fmt.Errorf("algorand service is not active after the upgrade: %w", err)
					}
					client, err := journeyClient(ctx)
					if err != nil {
						return err
					}
					if err := statewalker.WaitFor(ctx, client, statewalker.ExpectedState{State: algod.StableState}, stageTimeout); err != nil {
						return err
					}
					if err := statewalker.WaitForRoundsToAdvance(ctx, client, 2, stageTimeout); err != nil {
						return err
					}
					// The registered key must have survived the upgrade intact
					return waitForOnline(ctx, client, userAddress, voteLastValid, stageTimeout)
				},
			},
			{
				Name: "partkey-regen",
				Run: func(ctx context.Context, env *statewalker.Env) error {
					client, err := journeyClient(ctx)
					if err != nil {
						return err
					}
					threshold := voteLastValid - journeyRegenMargin
					log.Info("Waiting for the key to near its expiry", "threshold", threshold, "expiry", voteLastValid)
					if err := statewalker.WaitFor(ctx, client, statewalker.ExpectedState{MinRound: threshold}, regenTimeout); err != nil {
						return err
					}
					status, _, err := statewalker.CurrentStatus(ctx, client)
					if err != nil {
						return err
					}
					newLastValid := status.LastRound + 2_000_000
					log.Info("Renewing the participation key", "newLastValid", newLastValid)
					if out, err := journeyExec(journeyGoal(fmt.Sprintf(
						"account renewpartkey -a %s --roundLastValid %d -d %s",
						userAddress, newLastValid, dataDir()))); err != nil {
						return fmt.Errorf("partkey renewal failed: %w\n%s", err, out)
					}
					return nil
				},
				Verify: func(ctx context.Context, env *statewalker.Env) error {
					client, err := journeyClient(ctx)
					if err != nil {
						return err
					}
					// The account must be online with a key expiring after the old
					// one, proving the regenerated key superseded it
					deadline := time.Now().Add(stageTimeout)
					for {
						account, err := algod.GetAccount(client, userAddress)
						if err == nil && account.Status == "Online" && account.Participation != nil &&
							uint64(account.Participation.VoteLastValid) > voteLastValid {
							log.Info("Regenerated key is registered", "oldExpiry", voteLastValid, "newExpiry", account.Participation.VoteLastValid)
							return nil
						}
						if time.Now().After(deadline) {
							return fmt.Errorf("the regenerated key did not supersede the old one (account: %+v)", account)
						}
						time.Sleep(2 * time.Second)
					}
				},
			},
		},
	}
}

// waitForBalance polls the account until it holds at least the given amount.
func waitForBalance(ctx context.Context, client api.ClientWithResponsesInterface, address string, amount uint64, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		account, err := algod.GetAccount(client, address)
		if err == nil && uint64(account.Amount) >= amount {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("account %s was not funded within %s", address, timeout)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}

// waitForOnline polls the account until it is online; when expectedLastValid
// is non-zero, the registered key must also expire exactly at that round.
func waitForOnline(ctx context.Context, client api.ClientWithResponsesInterface, address string, expectedLastValid uint64, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		account, err := algod.GetAccount(client, address)
		if err == nil && account.Status == "Online" && account.Participation != nil {
			if expectedLastValid == 0 || uint64(account.Participation.VoteLastValid) == expectedLastValid {
				return nil
			}
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("account %s did not come online with the expected key within %s (status: %+v)", address, timeout, account.Status)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}

func init() {
	stageCmd.Flags().StringVar(&nodekitOldTag, "nodekit-old", "v1.6.0", "pinned older nodekit release tag the journey installs before upgrading")
	stageCmd.Flags().StringVar(&algodOldVersion, "algod-old", "", "pinned older algorand apt package version (empty installs the latest)")
}
