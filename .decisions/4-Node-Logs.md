# ℹ️ Overview

Reading a node's logs required knowing that algod writes a logrus JSON file, where it puts it, and
reaching for `tail -f | jq`. The `logs` command makes that a first-class operation. The decisions below
are the ones most likely to be revisited, recorded so they are not re-argued from scratch.

## ✅ Decisions

- **SHOULD** read the log file only, and **SHOULD NOT** fall back to journalctl, launchd or docker
- **SHOULD** resolve the log's location from `config.json` rather than assuming the data directory
- **SHOULD** show warnings and errors by default, with `--all` for every level
- **SHOULD** treat `--lines` as a count of matching entries, not of raw lines
- **SHOULD** show every match by default, with `--lines N` for the newest N
- **SHOULD** read the rotated archives as part of one history, with `--file` to read one file alone
- **SHOULD** match `--filter` as plain text against the message, not as a regular expression
- **SHOULD** say on stderr, before the first entry, which files are being read and what is being hidden
- **SHOULD** poll while following, rather than watching the file with inotify/kqueue
- **SHOULD** emit the original log lines verbatim for `--json`

## 🔨 Deliverables

- `nodekit logs` with `--follow`, `--lines`, `--level`, `--all`, `--since`, `--filter`, `--json` and `--file`
- A reusable reader in `internal/algod/logs` that is independent of the CLI

## 💬 Rationale

**File only.** The log file is structured JSON on every platform, so filtering by level is exact and the
output is identical everywhere. A service-manager fallback would give three different formats and three
different failure modes. The gap it would cover, a node that dies before writing anything, is better
served by telling the user where that output went: `LogSizeLimit: 0` is detected and reported, pointing
at the journal on Linux and `/tmp/algod.out` on macOS.

**Resolve, do not assume.** algod's `ResolveLogPaths` puts the live log in the data directory, or in
`HotDataDir`, or in `LogFileDir`, with the last of those winning. `config.Config.LogPaths` mirrors that
function so the two cannot drift. Assuming `<datadir>/node.log` silently reads the wrong file on any node
that uses a separate hot disk.

**Warnings and errors by default, counting matches.** On a healthy node warnings are well under one
percent of the file, so a command that filtered the last N lines would almost always print nothing.
Counting matches is what makes the default invocation useful, and it is what forces `--lines N` to walk
backwards from the end rather than scan forwards.

**Every match by default.** A count has to be filled from the newest end, which is what the backward walk is
for, but it also caps the history at whatever the walk could reach: a 512 MiB budget stops a search that
matches nothing from reading a whole gigabyte log, and it stopped `--lines 0` short of the start of the file
as well. Since warnings are well under one percent of a real log, the complete set of them is a readable
amount of output and the right default. Asking for everything is served by streaming forwards from the
oldest source instead of walking backwards: the same bytes are read, but entries print as they are found
and none are held. On a 692 MiB log that is a million entries at 58 MiB of memory, against 4.7 GiB for the
same request buffered. The budget still applies to `--lines N`, which is the search that can fail to find
what it is looking for.

**Archives are part of the history.** The live log only goes back to the last rotation, which on a busy
node is hours. The archive path was already resolved, so reading it is the difference between "the last
hour" and "everything algod still has". There is no flag to turn that off: the reason to want one is to
read a single file, which `--file` says directly and also covers the log that was copied off a machine.
`LogArchiveName` is a Go template, so a node keeping dated archives has several at once and the resolved
path is a pattern: those are found by globbing and ordered by modification time, since the names sort in
whatever order the template gives. algod hands the rotated file to `gzip` or `bzip2` when the name says
so, and those are decompressed on the way past. A compressed archive cannot be walked from its end, so
`--lines N` reads it forwards and keeps the newest N matches in a ring.

**Plain text, not patterns.** A log line is dense with characters a regular expression would claim:
`round=48291043`, `1.2.3.4:4160`, `[::1]`, `error: connection reset (*Service).mainLoop`. A search that
silently reinterprets those is a worse default than one that cannot express alternation, and anyone who
wants alternation still has a pipe. Taking the text at face value also means the same bytes can be
looked for in the undecoded line, so `--filter` gets the prescreen's full speedup rather than only the
subset of patterns that reduce to one literal.

**Say what is being hidden.** The default view is a filtered one, and it prints nothing to distinguish
"your node logged no warnings" from "you asked for warnings". A single stderr line naming the files
being read and the filters in force settles that before the first entry, and it is on stderr so a pipe
still sees only log entries. Naming the files is also the only way the archives being read is visible at
all.

**Polling over fsnotify.** Both backends do report the rename: `IN_MOVE_SELF` on Linux, `NOTE_RENAME` on
macOS. The event is not what is missing. It says the old path left, not when algod recreates `node.log`,
so following across a rotation needs a second watch on the directory and still races the new file being
written before that watch exists. Truncation in place is not a rename at all — an operator's `>` or a
copy-truncate arrives as a write — and is caught only by comparing the file's size against the read
offset. That comparison is a periodic `stat`, so a watch would sit on top of the loop already doing the
work, to save a quarter second of latency on a stream a human is reading. This is what `tail -F` does.

**Verbatim JSON.** Passing the original bytes through means `nodekit logs --json | jq` and reading the
file directly produce the same objects, with no field reordering and nothing lost to a round trip. Lines
that were never JSON are wrapped so the stream stays valid NDJSON.

## 🚧 Deliberately deferred

- Stopping the backward walk early on `--since`. It cannot be done without changing what `--since` means:
  `Keep` deliberately never drops an entry with no usable timestamp, so a panic dump older than the bound is
  shown today, and an early exit would silently stop showing it
- Showing the date on entries older than today; reading across a rotation makes a bare `15:04:05` ambiguous
- A logs page in the TUI; `Follow` is context-driven and already has the right shape to feed one
- Generalising `--json` into an `--output` flag shared by other commands
