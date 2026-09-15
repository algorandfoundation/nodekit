# Review: PR #193 - feat: nodekit logs

**Recommendation: approve with comments**

## Scope
New `cmd/logs.go` plus an `internal/algod/logs` package: reads `node.log`
(wherever `config.json` points it), filters by level (warn+ by default,
`--all` for everything), supports `--lines`, `--filter` (plain-text message
match), `--since`, `--json`, `--file`, rotated-archive history, and `--follow`
streaming. Extensive unit tests (entry parsing, filtering, archives, follow)
and a `.decisions/4-Node-Logs.md` design note.

## Verification
- `go build` + full `go test ./...` on the PR branch: pass.
- E2E against a statewalker private-network node:
  - Default: `reading node.log · showing: warn, error, fatal, panic` with only
    WARN entries rendered. ✔
  - `--all --lines 3`: newest 3 entries of any level. ✔
  - `--all --filter Catchpoint --lines 2`: plain-text filter matches. ✔
  - `--all --json --lines 1`: raw canonical JSON entry emitted. ✔
  - `--follow`: 2-line backlog then live streaming of new entries while the
    node produced blocks. ✔
  - **Stopped node** (`goal node stop`): `nodekit logs` still reads the full
    history; the deliberate absence of a "must be running" PreRun works as
    designed and is the command's main use case. ✔

## Code notes / suggestions
1. **PR contains PR #191's commits** (`cmd/restart.go`, `man/nodekit_restart.md`
   appear in this diff). Rebase after #191 merges (or mark as stacked) so the
   review surface is only the logs feature.
2. **Large diff (~6k lines)**: much of it is `api/lf.go`/`api/status.go`
   regeneration and `codegen-overlay.yaml`. A note in the PR description about
   which files are generated would help reviewers; consider splitting codegen
   churn from the feature commit.
3. The `-f` = `--follow` / `-F` = `--file` flag choice (avoiding the global
   `--force` rebinding) is well documented in the code comment. Good.
4. Behavior around unparseable lines ("crash output is always shown") verified
   in unit tests; nothing further needed.
