# Review: PR #192 - fix: track algod spec instead of pinning a stale one

**Recommendation: approve with comments**

## Scope
`make generate` now resolves the latest `-stable` go-algorand release from the
GitHub API (overridable via `ALGOD_VERSION=`) instead of pinning
`v3.26.0-stable`. The regenerated client (spec v5.0.1-stable) changes many
fields from `int` to `uint64` and makes state-dependent fields optional
pointers. The nodekit side migrates accordingly: underflow clamps in
`GetBlockMetrics`, `formatScheduledUpgrade`, `formatProtocolVote` and
`GetExpiresTime`, a `deref()` helper for optional fields in `Status.Merge`,
and a zero-`RoundTime` guard in `GenerateCmd`. A `codegen-overlay.yaml` maps
the spec's `x-go-type: basics.*` annotations (which ship without imports and
would not compile) to plain Go types. Branch is stacked on PR #191
(`feat/restart-command`) and also carries `fix: make --force reach the
platform stop`.

## Verification

### Codegen
- Unpinned `make generate` resolved `v5.0.1-stable` and left the committed
  `api/` byte-identical (`git status` clean); same with
  `ALGOD_VERSION=v5.0.1-stable`. Reproducible. ✔
- Version-sort edge case: fed the resolver pipeline a late backport tag
  (`v4.7.5-stable` listed after `v5.0.1-stable`); `sort -V | tail -1` still
  picks `v5.0.1-stable`. `per_page=100` keeps older lines from crowding out
  the newest. ✔
- `codegen-overlay.yaml`: the five `basics.*` mappings match what the types
  wrap upstream (`Round`/`AppIndex`/`AssetIndex`/`Micros` are uint64,
  `Address` is a string in the API). `strict: false` is correctly justified
  for older specs. The comment warning not to blanket-strip `x-go-type` is a
  good guardrail. ✔
- `go build ./...`, `go vet ./...`, full `go test ./...` (16 packages): pass. ✔

### Live network (statewalker private network, fresh chain)
- **Young-chain metrics clamp**: on a chain at round 5-60 (below the 100-round
  metrics window) the TUI renders `Round time: 0.00s`, `TPS: 0.00` and
  `Expires: N/A`; no ~1.8e19 garbage and no crash. Direct probe:
  `GetBlockMetrics(round=511, window=100000)` returns the `invalid window`
  error instead of requesting round `round-window` underflowed; with
  `window=100` it returns sane values (AvgTime 2.78s). ✔
- **`GenerateCmd` zero-`RoundTime`**: returns `round time is not yet known,
  please wait until your node is fully synced` instead of dividing by zero. ✔
- **`Status.Merge` deref (fast catchup)**: ran `walk fast-catchup` three
  times; `WaitFor` observed `FAST-CATCHUP` through the same `Merge` path the
  TUI uses, and the node returned to `RUNNING` with rounds advancing. The
  transitional payload (catchpoint reported before any counters) completes in
  milliseconds on a small ledger, so it was additionally verified by feeding
  `Merge` the exact payload shape algod emits at catchup start
  (`{"catchpoint":"400#..."}`, all counters omitted): result is
  `State=FAST-CATCHUP` with zero counters, no nil dereference. ✔
- **Optional vote fields**: `Merge` of a payload with
  `upgrade-next-protocol-vote-before` present but every vote counter omitted
  yields zeros (algod omits the denominator when no vote is in progress). ✔
- **Live upgrade voting**: `statewalker upgrade vote --vote-rounds 300` on a
  fresh chain; TUI renders `Protocol Upgrade: Voting 2% complete, 100% Yes`
  from the deref'd fields. ✔
- **`formatProtocolVote` fail-threshold clamp**: with
  `UpgradeVotesRequired=150 > UpgradeVoteRounds=100` the view renders
  `Voting 40% complete, 75% No, will fail`; pre-clamp the threshold would
  wrap to ~1.8e19 and silently suppress `will fail`. ✔
- **`GetExpiresTime`**: live accounts page shows sane dates (`11 Dec 26` for
  genesis keys with ~3M-round validity at the measured 2.5s round time),
  consistent with the PR #195 review results; a `voteLastValid` behind
  `LastRound` clamps to "now" instead of year ~294247. ✔

## Code notes / suggestions
1. **No record of the generated spec version.** `make generate` tracks a
   moving target, so the committed `api/` cannot be traced to the spec that
   produced it (this review had to re-run generation to confirm it matches
   v5.0.1-stable). Consider writing the resolved version to a small file
   (e.g. `api/.algod-version`) or a header comment during generation, so
   drift is visible in review diffs.
2. **Unauthenticated GitHub API resolution.** The releases endpoint is rate
   limited (60/hr per IP); CI or busy dev machines can hit it and fail
   `make generate`. The `ALGOD_VERSION` pin is the escape hatch; worth a
   mention in CONTRIBUTING, and optionally `-H "Authorization: ..."` when
   `GITHUB_TOKEN` is set.
3. **Stacked on PR #191.** `feat/restart-command` commits are in this diff;
   rebase (or merge #191 first) so the review surface is only the spec
   tracking. The included `fix: make --force reach the platform stop`
   addresses the macOS half of note 1 in pr-191-restart.md (now updated);
   consider folding that commit into #191 itself.
4. The clamp comments (`ui/protocol.go`, `internal/algod/block.go`,
   `deref()` in `status.go`) explain *why* each guard exists, which made this
   review much easier. Good practice worth keeping.
5. `internal/test/client.go` mock updates mirror the pointer migration
   faithfully; `Status.Merge` unit tests cover the new zero paths.
