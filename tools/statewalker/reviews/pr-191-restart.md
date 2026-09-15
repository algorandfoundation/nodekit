# Review: PR #191 - feat: nodekit restart

**Recommendation: approve with comments**

## Scope
New `cmd/restart.go` (`nodekit restart [-f/--force]`): stop the algod service,
wait `StopTimeout`, verify it stopped, start it again. Extracts the start/stop
messages into shared constants, threads the `force` flag through
`algod.Stop(force)`, adds `NotRunningStartErrorMsg`, and ships a man page.

## Verification
- `go build` + full `go test ./...` on the PR branch: pass.
- E2E against a statewalker private-network node (goal-managed, i.e. **not**
  installed as a system service):
  - `nodekit restart -d <Primary>` prints `Stopping Algod` then `FATA failed to stop
    Algod`. Expected: the command manages the system service (systemd/launchd),
    and there is none here. The node itself was untouched. ✔ (safe failure)
  - `nodekit restart --force -d <Primary>` gives an identical result on Linux.
- Happy-path (service-managed algod) exercised in the PR #197 container review;
  see pr-197-upgrade-order.md.

## Code notes / suggestions
1. **`--force` is a no-op on Linux.** `algod.Stop(force)` only forwards `force`
   to `mac.Stop(force)`; `linux.Stop()` ignores it. Combined with
   `NeedsToBeRunningToRestart` skipping *all* preflight when `--force` is set,
   on Linux the flag only disables the safety checks without changing the stop
   behavior. Worth documenting in the flag help or wiring force into the Linux
   path for symmetry.
   **Update (PR #192 review):** the stacked branch `fix/track-latest-algod-spec`
   carries `fix: make --force reach the platform stop` (commit `7890166`),
   which threads `force` into `algod.Stop(force)` and lets it bypass
   the `MustBeServiceMsg` guard in `mac.Stop`. That addresses the macOS half of
   this note. On Linux `linux.Stop()` still takes no `force` (there is no
   service guard to bypass; `systemctl stop` is always attempted), so the flag
   help/doc suggestion stands. Re-verified on this session's private network:
   `restart --force` behavior on Linux is unchanged.
2. **Fixed sleep + single check.** `time.Sleep(StopTimeout)` followed by one
   `IsRunning` probe can mis-report a slow shutdown as failure (and a fast
   shutdown always pays the full sleep). Polling `IsRunning` up to a deadline
   would be more robust; fine as a follow-up.
3. Error message when there is no service: `failed to stop Algod` does not say
   *why*. A hint that restart requires a service-managed node (like the
   explanations used elsewhere) would save confusion for goal-started nodes.
4. Reuse of the package-global `force` variable shared with stop/install is
   pre-existing practice, and the restart command rebinding it is consistent
   with `stop`; no action needed.
