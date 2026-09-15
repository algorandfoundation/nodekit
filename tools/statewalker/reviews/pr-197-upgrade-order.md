# Review: PR #197 - fix: re-exec updated nodekit before algod upgrade (IS #190)

**Recommendation: request changes** (real bug found in the re-exec path)

## Scope
`nodekit upgrade` now self-upgrades NodeKit first, then re-executes the new
binary (marked via `NODEKIT_REEXEC=1`) so the *updated* NodeKit performs the
algod upgrade. `system.Reexec`/`system.IsReexec` added with unit tests;
`IsReexec` prevents an upgrade→re-exec loop.

## Verification
- `go build` + full `go test ./...` on the PR branch: pass.
- E2E in a throwaway `ubuntu:24.04` container, using a build that mirrors a
  production tagged release exactly (CD.yaml recipe: `GOOS=linux GOARCH=amd64
  CGO_ENABLED=0 go build -ldflags "-X main.version=1.6.0" -o
  nodekit-amd64-linux *.go`), installed at `/usr/bin/nodekit`. This is the real
  upgrade path: `NeedsUpgrade` is true because the stamped `1.6.0` differs from
  the actual latest release (`v1.6.3`), and `system.Upgrade` downloads the real
  production `v1.6.3` binary from GitHub releases. Ran `nodekit upgrade`:
  ```
  INFO Upgrading NodeKit
  DEBU fetching .../releases/latest/download/nodekit-amd64-linux
  DEBU backing up to /usr/bin/.nodekit.bak
  DEBU deploying /usr/bin/.nodekit.tmp to /usr/bin/nodekit
  INFO NodeKit upgraded successfully. Restarting with the latest version.
  FATA fork/exec /usr/bin/.nodekit.bak: no such file or directory
  ```
  The self-upgrade succeeded (`nodekit --version` reports `1.6.3` afterwards,
  `.nodekit.bak` is gone), but the re-exec **failed** with exit code 1, and the
  algod upgrade never ran. The regression this PR sets out to fix (upgrade
  stops before touching algod) becomes a hard failure. Reproduced twice, on
  independent runs/builds (`1.0.0` and `1.6.0` stamps).
- Counterfactual check (confirms the diagnosis, not just the symptom): same
  container setup with a one-line variant of the PR that captures
  `os.Executable()` *before* `system.Upgrade` and passes it to `Reexec`. The
  run then proceeds past the re-exec: `NodeKit upgraded successfully` →
  `Upgrading Algod` (which fails only because algod isn't installed in the
  container, as expected). The self-upgrade → re-exec → algod-upgrade ordering
  works end-to-end once the path is captured early, and no re-exec loop occurs.

## Bug
`Reexec` calls `os.Executable()` *after* `system.Upgrade` replaced the binary.
On Linux `os.Executable()` resolves `/proc/self/exe`, which follows the running
inode: `Upgrade` renames the live binary to `.nodekit.bak`, deploys the new
one, then **deletes the backup**, so `os.Executable()` returns the now-deleted
`.nodekit.bak` path and `fork/exec` fails. (Even if the backup survived, it
would re-exec the *old* binary.)

**Fix direction:** capture the executable path *before* the upgrade (or return
the deployed path from `system.Upgrade`) and exec that path. The unit tests
pass because they exercise `reexecCommand` with a caller-supplied path, an
integration gap that this e2e run exposes.

## Additional notes
0. Real-world impact window: current production (`v1.6.3`) does not contain the
   re-exec code, so the first release that ships this PR upgrades cleanly *to*
   it. The failure hits every user running that first release (or any later
   one with this code) when they upgrade to the *next* release — i.e. it is a
   time bomb for the release after this PR ships, on every Linux install.
1. `Reexec` uses `cmd.Start()` and the parent returns immediately: the parent
   exits while the child still owns the terminal (and may prompt for sudo).
   Prefer `syscall.Exec` on unix (nodekit only supports linux/darwin) or at
   least `cmd.Run()` + `os.Exit` with the child's exit code.
2. Ordering itself (NodeKit first, then algod) is correct and verified: with a
   dev build (`NeedsUpgrade=false`) the command goes straight to the algod
   step, and the failed run above shows the NodeKit step strictly precedes it.
3. Version-check dependency: `NeedsUpgrade` is only true when the build was
   stamped via ldflags; `@dev` builds skip the new path entirely (fine).
