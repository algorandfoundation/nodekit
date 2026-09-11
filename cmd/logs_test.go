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

// runLogs invokes the command the way a user would, returning what it wrote to
// stderr, where every explanation of an empty result goes.
func runLogs(t *testing.T, args ...string) string {
	t.Helper()

	var out, errOut bytes.Buffer
	logsCmd.SetOut(&out)
	logsCmd.SetErr(&errOut)
	t.Cleanup(func() {
		logsCmd.SetOut(nil)
		logsCmd.SetErr(nil)
		logsCmd.Flags().Visit(func(f *pflag.Flag) { f.Changed = false })
		logsDataDir, logsSince, logsFile, logsLevel, logsFilter = "", "", "", "", ""
		logsFollow, logsAll, logsJSON, logsLines = false, false, false, defaultLogLines
	})

	// RunE rather than Execute: Execute runs from the root command, which opens
	// a TTY this test has no use for and no access to.
	require.NoError(t, logsCmd.ParseFlags(args))
	require.NoError(t, logsCmd.RunE(logsCmd, nil))
	return errOut.String()
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
