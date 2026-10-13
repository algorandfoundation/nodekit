package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/algorandfoundation/nodekit/cmd/utils/explanations"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// nodeWithArchivedHistory builds a data directory holding an empty live log and
// one archive old enough for --since to prune, which is the shape that made the
// command claim the node had never logged anything.
func nodeWithArchivedHistory(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "genesis.json"), []byte(`{}`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"LogSizeLimit":1073741824}`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "node.log"), nil, 0o600))

	archive := filepath.Join(dir, "node.archive.log")
	entry := fmt.Sprintf(`{"level":"warning","msg":"older history","time":%q}`, time.Now().Add(-48*time.Hour).Format(time.RFC3339))
	require.NoError(t, os.WriteFile(archive, []byte(entry+"\n"), 0o600))

	old := time.Now().Add(-48 * time.Hour)
	require.NoError(t, os.Chtimes(archive, old, old))

	return dir
}

func resetLogsFlags() {
	logsCmd.Flags().Visit(func(f *pflag.Flag) { f.Changed = false })
	logsDataDir, logsSince, logsFile, logsLevel, logsFilter = "", "", "", "", ""
	logsFollow, logsAll, logsJSON, logsLines = false, false, false, defaultLogLines
}

// runLogs invokes the command the way a user would, returning what it wrote to
// stderr, where every explanation of an empty result goes.
func runLogs(t *testing.T, args ...string) string {
	t.Helper()
	_, errOut := runLogsStreams(t, context.Background(), args...)
	return errOut
}

// runLogsStreams is runLogs with the entries as well as the explanations, and a
// context of the caller's choosing.
//
// An already cancelled context is how --follow is exercised here: Follow drains
// what the scan left and then returns on the first look at Done, so the backlog
// is written and the command comes back instead of polling for the life of the
// test.
func runLogsStreams(t *testing.T, ctx context.Context, args ...string) (string, string) {
	t.Helper()

	var out, errOut bytes.Buffer
	logsCmd.SetOut(&out)
	logsCmd.SetErr(&errOut)
	logsCmd.SetContext(ctx)

	// Flags are package state shared by every invocation, so they are cleared
	// before each one and not only at the end of the test.
	resetLogsFlags()
	t.Cleanup(func() {
		logsCmd.SetOut(nil)
		logsCmd.SetErr(nil)
		logsCmd.SetContext(context.Background())
		resetLogsFlags()
	})

	// RunE rather than Execute: Execute runs from the root command, which opens
	// a TTY this test has no use for and no access to.
	require.NoError(t, logsCmd.ParseFlags(args))
	require.NoError(t, logsCmd.RunE(logsCmd, nil))
	return out.String(), errOut.String()
}

// nodeWithFloor builds a data directory for a node configured to log at the
// given BaseLoggerDebugLevel, which is what decides whether the view is
// genuinely incomplete.
func nodeWithFloor(t *testing.T, floor uint32) string {
	t.Helper()
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "genesis.json"), []byte(`{}`), 0o600))
	config := fmt.Sprintf(`{"LogSizeLimit":1073741824,"BaseLoggerDebugLevel":%d}`, floor)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.json"), []byte(config), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "node.log"), nil, 0o600))

	return dir
}

// --all names no level: it asks for whatever the log holds. Measuring it
// against the node's floor fired this on every default node, because --all asks
// for trace and no algod writes one.
func TestLogsDoesNotWarnAboutTheFloorWhenAllNamesNoLevel(t *testing.T) {
	const warning = "lower levels are never written to the log"

	assert.NotContains(t, runLogs(t, "--datadir", nodeWithFloor(t, 5), "--all"), warning,
		"a node at debug writes everything algod can")
	assert.NotContains(t, runLogs(t, "--datadir", nodeWithFloor(t, 4), "--all"), warning,
		"info is algod's default, so this fired on very nearly every node")
	assert.NotContains(t, runLogs(t, "--datadir", nodeWithFloor(t, 1), "--all"), warning,
		"even here --all shows all there is; an empty result is explained on its own")

	// A level the user named and the node never writes is what the warning is
	// for, and still reaches them.
	assert.Contains(t, runLogs(t, "--datadir", nodeWithFloor(t, 4), "--level", "debug"), warning,
		"--level debug against an info floor asks for entries that do not exist")
}

func TestLogsDoesNotCallAnArchivedHistoryAnEmptyLog(t *testing.T) {
	dir := nodeWithArchivedHistory(t)

	// --since is newer than the archive, so PruneSources drops it before the
	// scan opens anything and only the empty live log is left to read. That is
	// a search with no matches, not a node that has never written a line.
	got := runLogs(t, "--datadir", dir, "--since", "1h")

	assert.NotContains(t, got, explanations.LogsEmptyMsg,
		"the node has a rotated archive; --since pruning it away does not unmake it")
	assert.Contains(t, got, explanations.LogsNoMatchMsg)
}

func TestLogsCallsALogWithNoHistoryAtAllEmpty(t *testing.T) {
	dir := nodeWithArchivedHistory(t)
	require.NoError(t, os.Remove(filepath.Join(dir, "node.archive.log")))

	// An empty live log and nothing behind it really is a node that has not
	// logged anything yet, and still has to say so.
	assert.Contains(t, runLogs(t, "--datadir", dir), explanations.LogsEmptyMsg)
}

// nodeWithBacklog builds a data directory whose live log holds more warnings
// than any fixed backlog default would have shown, so that a run stopping short
// of the oldest one is visible.
func nodeWithBacklog(t *testing.T, filler int) string {
	t.Helper()
	dir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(dir, "genesis.json"), []byte(`{}`), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "config.json"), []byte(`{"LogSizeLimit":1073741824}`), 0o600))

	start := time.Now().Add(-time.Hour)
	entry := func(i int, msg string) string {
		return fmt.Sprintf(`{"level":"warning","msg":%q,"time":%q}`,
			msg, start.Add(time.Duration(i)*time.Minute).Format(time.RFC3339))
	}

	lines := []string{entry(0, "oldest warning")}
	for i := 1; i <= filler; i++ {
		lines = append(lines, entry(i, fmt.Sprintf("filler warning %d", i)))
	}
	lines = append(lines, entry(filler+1, "newest warning"))

	log := filepath.Join(dir, "node.log")
	require.NoError(t, os.WriteFile(log, []byte(strings.Join(lines, "\n")+"\n"), 0o600))

	return dir
}

// --follow used to substitute a backlog of its own when --lines was not given,
// which left -n meaning one thing alone and another beside -f: `-n 0 -f` asked
// for the whole history where `tail -n 0 -f` asks for none of it. The default is
// the same in both now, and a stream that skips the history is spelled with
// --since.
func TestLogsFollowReplaysTheWholeHistoryByDefault(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	dir := nodeWithBacklog(t, 20)

	got, _ := runLogsStreams(t, ctx, "--datadir", dir, "--follow")
	assert.Contains(t, got, "oldest warning",
		"--follow with no --lines shows every match, the way the command does without it")
	assert.Contains(t, got, "newest warning")

	// --lines is still how the backlog is shortened.
	got, _ = runLogsStreams(t, ctx, "--datadir", dir, "--follow", "--lines", "2")
	assert.NotContains(t, got, "oldest warning", "--lines 2 asks for two entries")
	assert.Contains(t, got, "newest warning")

	// And --since is how the timestamped history is skipped.
	got, _ = runLogsStreams(t, ctx, "--datadir", dir, "--follow", "--since", "0s")
	assert.NotContains(t, got, "oldest warning")
	assert.NotContains(t, got, "newest warning", "every entry in the fixture is older than the bound")
}

// --since 0s skips the timestamped history but not a plain-text line near the
// end of the log: Keep never drops a line without a timestamp, and --since
// bounds the region read rather than each line in it. A crash is what such a
// line usually is, so the help text says it may still be shown rather than
// promising a stream that starts empty.
func TestLogsSinceNowStillShowsPlainTextNearTheEnd(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	dir := nodeWithBacklog(t, 20)
	log, err := os.OpenFile(filepath.Join(dir, "node.log"), os.O_APPEND|os.O_WRONLY, 0o600)
	require.NoError(t, err)
	_, err = log.WriteString("panic: runtime error: invalid memory address or nil pointer dereference\n" +
		fmt.Sprintf(`{"level":"warning","msg":"after the panic","time":%q}`, time.Now().Add(-time.Minute).Format(time.RFC3339)) + "\n")
	require.NoError(t, err)
	require.NoError(t, log.Close())

	for _, args := range [][]string{{"--since", "0s"}, {"--since", "0s", "--follow"}} {
		got, _ := runLogsStreams(t, ctx, append([]string{"--datadir", dir}, args...)...)
		assert.Contains(t, got, "panic: runtime error", "%v: a line with no timestamp is not dropped by --since", args)
		assert.NotContains(t, got, "newest warning", "%v", args)
		assert.NotContains(t, got, "after the panic", "%v", args)
	}
}
