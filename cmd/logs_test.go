package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
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

	var out, errOut bytes.Buffer
	logsCmd.SetOut(&out)
	logsCmd.SetErr(&errOut)

	// Flags are package state shared by every invocation, so they are cleared
	// before each one and not only at the end of the test.
	resetLogsFlags()
	t.Cleanup(func() {
		logsCmd.SetOut(nil)
		logsCmd.SetErr(nil)
		resetLogsFlags()
	})

	// RunE rather than Execute: Execute runs from the root command, which opens
	// a TTY this test has no use for and no access to.
	require.NoError(t, logsCmd.ParseFlags(args))
	require.NoError(t, logsCmd.RunE(logsCmd, nil))
	return errOut.String()
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
