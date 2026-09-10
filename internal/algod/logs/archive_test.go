package logs

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// An archive algod is part way through compressing ends early, and the whole
// point of reading the history best effort is that this costs its entries and
// nothing else.
func TestScanSkipsAnArchiveItCannotRead(t *testing.T) {
	dir := t.TempDir()
	live := writeAt(t, dir, "node.log", line("warning", "from the live log")+"\n")
	truncated := writeGzipped(t, dir, "node.archive.log.gz",
		line("warning", "from the archive")+"\n")
	truncate(t, truncated, 0.6)

	for name, lines := range map[string]int{"a count": 100, "everything": 0} {
		t.Run(name, func(t *testing.T) {
			got, result := collect(t, []string{live, truncated}, lines, Filter{MinLevel: LevelWarn})

			assert.Equal(t, []string{"from the live log"}, got)
			require.Len(t, result.Skipped, 1)
			assert.Equal(t, truncated, result.Skipped[0].Path)
			assert.Error(t, result.Skipped[0].Err)
		})
	}
}

// The live log is not best effort. Its failure is the answer to the command.
func TestScanFailsOnAnUnreadableLiveLog(t *testing.T) {
	dir := t.TempDir()
	live := writeGzipped(t, dir, "node.log.gz", line("warning", "unreachable")+"\n")
	truncate(t, live, 0.6)

	for name, lines := range map[string]int{"a count": 100, "everything": 0} {
		t.Run(name, func(t *testing.T) {
			_, err := Scan([]string{live}, lines, Filter{MinLevel: LevelWarn}, func(Entry) error { return nil })
			assert.Error(t, err)
		})
	}
}

// A compressed archive is not being appended to, and an offset into one counts
// decompressed bytes, so seeking to it would land anywhere at all.
func TestFollowRefusesACompressedArchive(t *testing.T) {
	path := writeGzipped(t, t.TempDir(), "node.archive.log.gz", line("warning", "old")+"\n")

	err := Follow(context.Background(), path, 0, Filter{}, func(Entry) error {
		return fmt.Errorf("nothing should have been emitted")
	}, FollowOptions{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "compressed archive")
}

// Decompressing is work like any other and comes out of the same budget, so an
// archive too large to finish stops the scan rather than running past the cap.
func TestCompressedArchiveIsBoundedByTheBudget(t *testing.T) {
	dir := t.TempDir()
	live := writeAt(t, dir, "node.log", line("warning", "from the live log")+"\n")

	var b strings.Builder
	for i := 0; i < 400; i++ {
		b.WriteString(line("warning", fmt.Sprintf("archived %d", i)) + "\n")
	}
	archive := writeGzipped(t, dir, "node.archive.log.gz", b.String())

	withMaxScanned(t, 1024) // smaller than the archive decompresses to
	got, result := collect(t, []string{live, archive}, 100, Filter{MinLevel: LevelWarn})

	assert.True(t, result.ScanLimited, "the answer is short of what was asked for")
	assert.Empty(t, result.Skipped, "running out of budget is not a failure to read")
	assert.Equal(t, []string{"from the live log"}, got,
		"an archive that could not be finished contributes nothing, rather than a hole in the history")
}

// A budget used up exactly as a file ends used to leave the rest of the history
// unread without anything saying so.
func TestBudgetSpentAtAFileBoundaryIsStillReported(t *testing.T) {
	dir := t.TempDir()

	var b strings.Builder
	for i := 0; i < 20; i++ {
		b.WriteString(line("info", fmt.Sprintf("noise %d", i)) + "\n")
	}
	live := writeAt(t, dir, "node.log", b.String())
	archive := writeAt(t, dir, "node.archive.log", line("warning", "wanted")+"\n")

	withMaxScanned(t, int64(b.Len())) // exactly the live log, to the byte
	got, result := collect(t, []string{live, archive}, 5, Filter{MinLevel: LevelWarn})

	assert.Empty(t, got)
	assert.True(t, result.ScanLimited, "the archive was never opened, which has to be said")
}

// truncate cuts a file to a fraction of its length, the way a rotated log looks
// while the compressor is still writing it.
func truncate(t *testing.T, path string, fraction float64) {
	t.Helper()
	info, err := os.Stat(path)
	require.NoError(t, err)
	require.NoError(t, os.Truncate(path, int64(float64(info.Size())*fraction)))
}
