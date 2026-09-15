# statewalker

A test harness that provisions Algorand networks and walks `algod` into the
states the nodekit TUI distinguishes (`RUNNING`, `SYNCING`, `FAST-CATCHUP`,
upgrade voting, participation-key lifecycles), so nodekit behavior can be
reviewed end to end. It is a standalone binary; nothing here ships in the
`nodekit` production binary.

## Requirements

- `goal`/`algod` on the `PATH` for `--mode private` (any recent release;
  developed against v5.x / go-algorand 4.x)
- `algokit` + Docker for `--mode localnet`

## Quick start

```sh
make statewalker             # builds bin/statewalker
make e2e-net-up              # private network: Relay + Primary + Node2 (anchor)
nodekit -d ~/.statewalker/net/private/Primary
make e2e-net-down            # stop; `bin/statewalker network down --delete` also removes it
```

`network up` prints the Primary data dir; every scenario targets **Primary**
only. `Node2` holds ~70% of the online stake and, together with the `Relay`,
is never stopped, wiped, or taken offline by any scenario, so the chain keeps
advancing no matter what a test does to Primary (stake split: Wallet1 20%
online + Wallet3 10% offline on Primary, Wallet2 70% online on Node2).

The relay is configured as archival with the ledger/block services enabled and
`CatchpointInterval: 100`, so catchpoint (fast) catchup works inside the
private network. The first catchpoint appears around round ~420
(`CatchpointLookback` 320 + interval).

## Commands

| Command | What it does |
|---|---|
| `network up --mode private\|localnet` | Provision + start; verifies `RUNNING` and that rounds advance |
| `network status` / `network down [--delete]` | Status / stop (optionally delete) |
| `partkeys seed --expiring 72h --expiring 720h [-a addr] [--online]` | Generate keys with wall-clock validity (converted via measured round time) |
| `partkeys online\|offline [-a addr]` | Register a Primary wallet on/offline; verifies the chain still advances |
| `partkeys zombie` | Install a key for the offline funded wallet without registering it (IS #41 sibling case) |
| `partkeys list` | `goal account partkeyinfo` for Primary |
| `walk syncing` | Stop Primary, wipe its ledger (partkeys preserved), restart → observe `SYNCING` → `RUNNING` |
| `walk fast-catchup` | Same, but starts catchpoint catchup from the anchor's latest label → observe `FAST-CATCHUP` |
| `traffic start\|stop [--interval 2s]` | Background self-payment loop (needed on localnet, where rounds only advance with traffic) |
| `upgrade vote [--vote-rounds N --threshold N --delay N]` | Install a `consensus.json` cloning the genesis protocol into an approved successor; restart; `/v2/status` reports vote rounds/yes-votes (the upgrade fully activates after the window) |
| `upgrade clear` | Remove the consensus overrides and restart |

All state assertions poll `/v2/status` through the same generated `api` client
and `internal/algod.Status` mapping the TUI uses, so a scenario "passing"
means nodekit would render that state.

## Staged scenarios

`statewalker stage <name> [--speed fast|real] [--keep] [--timeout 5m]`
composes the drivers above into named, reproducible scenarios. Each phase is
verified through `/v2/status` / `/v2/participation` before the next starts; on
failure the network (and journey container) is torn down unless `--keep` is
given, with container logs and the last status dumped first.

| Scenario | What it stages |
|---|---|
| `demo` | Seeded showcase for tape capture: long-validity online key, near-expiry key, zombie key, upgrade vote in progress. Stays up and prints `DATADIR=<dir>` for the tapes. |
| `e2e` | The PR fast suite: network up, partkeys seed/online/offline/zombie, syncing walk, upgrade vote, liveness invariant. Tears down when done. |
| `journey` | The full end-user lifecycle in a throwaway Debian/systemd container joined to the host private network: `install.sh` → pinned-old `nodekit install` (apt) → sync → fast catchup → go online → real `nodekit upgrade` + algod apt upgrade → partkey regeneration near expiry. Flags: `--nodekit-old vX.Y.Z`, `--algod-old <apt version>`. |

`--speed fast` clones the genesis consensus protocol with ~1s agreement
timeouts so catchup/expiry phases fit CI budgets; `--speed real` keeps the
stock MainNet-like cadence for tape realism.

The journey runs inside the container as a non-root `algo` user with a
`NOPASSWD:ALL` sudoers drop-in, mirroring a real end-user machine (nodekit
shells out to `sudo apt-get ...`). Every exec sets
`DEBIAN_FRONTEND=noninteractive` and is guarded by an inactivity watchdog that
kills and names any command that would stall on an interactive prompt, so
hangs surface as fast failures instead of 0%-CPU stalls. The orchestration
lives behind a small `Backend` interface (`internal/journey.go`), so a macOS
backend (brew/launchd on a mac runner) can be added later without touching the
scenario phases. Note: the journey's upgrade phase exercises the real release
channels; until PR #197 lands fixed, `nodekit upgrade` can hit the known
re-exec bug, which the scenario detects and reports while completing the algod
upgrade through the same apt path nodekit uses.

Makefile shortcuts: `make stage-demo`, `make e2e-fast` (PR suite), and
`make e2e-full` (adds the journey, `walk fast-catchup`, and the
localnet+traffic path).

## CI

`.github/workflows/e2e_test.yaml` runs exactly the local Makefile targets:

| Job | Trigger | Command |
|---|---|---|
| `fast` | every PR touching Go code, `ui/`, `.tapes/`, this tool or the Makefile | `make e2e-fast` |
| `full` | `run-full-e2e` PR label, `workflow_dispatch`, nightly schedule | `make e2e-full` |
| `tapes` | PRs touching `ui/`, `.tapes/` or `internal/` | `make tapes` + gif artifacts |

Add the `run-full-e2e` label to a PR when it touches long-running paths
(catchup, install, upgrade). Full-suite failures can come from upstream
release channels rather than the repo; the journey reports "upstream fetch"
errors distinctly. On merge to main, `.github/workflows/tapes_commit.yaml`
regenerates the gifs and auto-commits them when changed (bot identity,
`[skip tapes]` loop guard).

Localnet mode writes a shim data dir (`~/.statewalker/net/localnet`) with
`algod.net`/`algod.token` pointing at the algokit containers so `nodekit -d`
can attach; rounds only advance while `traffic start` is running. The stall
without traffic is expected localnet behavior.

## Capturing the TUI headlessly

```sh
(sleep 8; printf '\r'; sleep 3; printf '\r'; sleep 5) \
  | TERM=xterm-256color script -qec "stty cols 140 rows 40; nodekit -d <datadir>" tui.log
python3 -c "import pyte;s=pyte.Screen(140,40);st=pyte.Stream(s);st.feed(open('tui.log',errors='replace').read());print('\n'.join(l.rstrip() for l in s.display))"
```

## Relation to the go-algorand harness

How statewalker overlaps with the upstream go-algorand test harness
(netdeploy, fixtures, nettemplates) and why it reuses it through the `goal`
process boundary instead of Go imports is evaluated in
[`docs/upstream-overlap.md`](docs/upstream-overlap.md).

## Review notes

E2E review results for the open PRs/issues this harness was built to test live
in [`reviews/`](reviews/):

- [PR #195 / IS #128 - expiration datetime format](reviews/pr-195-expiration-format.md): **approve** (both formats verified in the TUI)
- [PR #191 - `nodekit restart`](reviews/pr-191-restart.md): **approve with comments** (`--force` is a no-op on Linux; fixed-sleep stop check)
- [PR #193 - `nodekit logs`](reviews/pr-193-logs.md): **approve with comments** (verified incl. `--follow` and stopped-node reads; contains PR #191 commits)
- [PR #197 / IS #190 - upgrade ordering](reviews/pr-197-upgrade-order.md): **request changes** (re-exec fails: `os.Executable()` resolves the deleted `.bak` after self-upgrade; reproduced in a container)
- [IS #41 - zombie key deletion](reviews/is-41-zombie-keys.md): **confirmed**, reproduction + fix direction documented
- [PR #192 - track algod spec](reviews/pr-192-track-spec.md): **approve with comments** (codegen reproducible at v5.0.1-stable; uint64 clamps and optional-field derefs verified live on a young chain, fast catchup and upgrade voting)
