package logs

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withSinceSearch shrinks the geometry of the bisect so that a small fixture
// exercises the same loop a gigabyte of node.log would.
func withSinceSearch(t *testing.T, probe, slack int64) {
	t.Helper()
	originalProbe, originalSlack := probeSize, sinceSlack
	probeSize, sinceSlack = probe, slack
	t.Cleanup(func() { probeSize, sinceSlack = originalProbe, originalSlack })
}

// lineAt is line() with a timestamp of its own, for fixtures that have to span
// a stretch of time rather than a single instant.
func lineAt(level, msg string, ts time.Time) string {
	return fmt.Sprintf(`{"level":%q,"msg":%q,"time":%q}`, level, msg, ts.Format(time.RFC3339Nano))
}

// timedLog writes count lines one minute apart, the last of them at end, and
// returns the content along with the byte offset each line starts at.
func timedLog(level string, count int, end time.Time) (string, []int64, []time.Time) {
	var (
		b       strings.Builder
		offsets = make([]int64, count)
		times   = make([]time.Time, count)
	)
	for i := 0; i < count; i++ {
		ts := end.Add(-time.Duration(count-1-i) * time.Minute)
		offsets[i] = int64(b.Len())
		times[i] = ts
		b.WriteString(lineAt(level, fmt.Sprintf("entry %d", i), ts) + "\n")
	}
	return b.String(), offsets, times
}

func TestPeekTimeReadsTheTopLevelTimestamp(t *testing.T) {
	ts := time.Date(2026, 9, 9, 14, 16, 19, 250000000, time.UTC)

	got, ok := peekTime([]byte(lineAt("warning", "hello", ts)))

	require.True(t, ok)
	assert.True(t, got.Equal(ts), "expected %s, got %s", ts, got)
}

// The whole value of peekTime is that a false answer is never a wrong one: the
// walk that uses it would stop early on a timestamp it should not have believed.
func TestPeekTimeDeclinesWhatItCannotBeSureOf(t *testing.T) {
	for name, line := range map[string]string{
		"a plain line":         "panic: runtime error: index out of range",
		"no time at all":       `{"level":"warning","msg":"x"}`,
		"a time only inside":   `{"details":{"time":"2026-09-09T14:16:19+02:00"},"level":"warning","msg":"x"}`,
		"an unreadable time":   `{"level":"warning","msg":"x","time":"yesterday"}`,
		"a line cut off early": `{"level":"warning","msg":"x","time":"2026-09-09T14:16:19+02:00"`,
	} {
		t.Run(name, func(t *testing.T) {
			_, ok := peekTime([]byte(line))
			assert.False(t, ok)
		})
	}
}

// A backward walk with a lower bound reads the window it was asked for and
// stops, rather than continuing to the budget over entries that cannot match.
func TestTailStopsAtTheSinceBoundary(t *testing.T) {
	withChunkSize(t, 256)

	end := time.Now().Truncate(time.Second)
	content, _, times := timedLog("warning", 200, end)
	path := writeLog(t, content)

	f := Filter{MinLevel: LevelWarn, Since: times[195]}
	found, scanned, reason, err := tailFile(path, -1, 100, f, newPrescreen(f), maxScanned)
	require.NoError(t, err)

	assert.Equal(t, stopSince, reason)
	assert.Len(t, found, 5, "only the entries inside the window")
	assert.Less(t, scanned, int64(len(content))/2, "stopped well short of the whole file")
}

// Reaching the lower bound is a complete answer, not a truncated one, so it must
// not raise the warning that sends the user back to search even more of the log.
func TestScanWithSinceDoesNotReportBeingLimited(t *testing.T) {
	withChunkSize(t, 256)
	withMaxScanned(t, 4096)

	end := time.Now().Truncate(time.Second)
	content, _, times := timedLog("warning", 200, end)
	dir := t.TempDir()
	path := writeAt(t, dir, "node.log", content)

	got, result := collect(t, []string{path}, 100, Filter{MinLevel: LevelWarn, Since: times[195]})

	assert.False(t, result.ScanLimited, "the walk reached the bound, not the budget")
	assert.Len(t, got, 5)
	assert.Equal(t, "entry 195", got[0], "oldest first")
	assert.Equal(t, "entry 199", got[len(got)-1])
}

// Without a bound the same fixture does exhaust the budget, which is what the
// warning is for. This is the control for the test above.
func TestScanWithoutSinceStillReportsTheBudget(t *testing.T) {
	withChunkSize(t, 256)
	withMaxScanned(t, 4096)

	end := time.Now().Truncate(time.Second)
	content, _, _ := timedLog("warning", 200, end)
	dir := t.TempDir()
	path := writeAt(t, dir, "node.log", content)

	_, result := collect(t, []string{path}, 100, Filter{MinLevel: LevelWarn})

	assert.True(t, result.ScanLimited)
}

func TestSinceOffsetStartsBeforeTheFirstEntryInRange(t *testing.T) {
	withSinceSearch(t, 256, 512)

	end := time.Now().Truncate(time.Second)
	content, offsets, times := timedLog("info", 400, end)
	path := writeLog(t, content)

	got, err := sinceOffset(path, -1, times[300])
	require.NoError(t, err)

	assert.Greater(t, got, int64(0), "the search should have skipped part of the file")
	assert.LessOrEqual(t, got, offsets[300], "starting here would miss entries in range")
}

// The offset is only ever a place to start reading, so anything the search
// cannot answer for has to come back as the beginning of the file.
func TestSinceOffsetFallsBackToTheStartOfTheFile(t *testing.T) {
	withSinceSearch(t, 256, 512)

	end := time.Now().Truncate(time.Second)
	content, _, times := timedLog("info", 400, end)
	dir := t.TempDir()

	t.Run("with no lower bound", func(t *testing.T) {
		got, err := sinceOffset(writeAt(t, dir, "node.log", content), -1, time.Time{})
		require.NoError(t, err)
		assert.Zero(t, got)
	})

	t.Run("when the whole file is in range", func(t *testing.T) {
		got, err := sinceOffset(writeAt(t, dir, "in-range.log", content), -1, times[0].Add(-time.Hour))
		require.NoError(t, err)
		assert.Zero(t, got)
	})

	t.Run("on a compressed archive that cannot be seeked into", func(t *testing.T) {
		got, err := sinceOffset(writeGzipped(t, dir, "node.archive.log.gz", content), -1, times[300])
		require.NoError(t, err)
		assert.Zero(t, got)
	})

	t.Run("on a file smaller than the slack it would rewind by", func(t *testing.T) {
		short, _, shortTimes := timedLog("info", 4, end)
		got, err := sinceOffset(writeAt(t, dir, "short.log", short), -1, shortTimes[3])
		require.NoError(t, err)
		assert.Zero(t, got)
	})
}

// The forward stream must emit the same entries whether or not it skipped ahead
// to reach them.
func TestSinceOffsetDoesNotChangeWhatIsEmitted(t *testing.T) {
	end := time.Now().Truncate(time.Second)
	content, _, times := timedLog("warning", 400, end)
	dir := t.TempDir()
	path := writeAt(t, dir, "node.log", content)
	filter := Filter{MinLevel: LevelWarn, Since: times[250]}

	withSinceSearch(t, 1<<30, 1<<30) // large enough that the search declines
	whole, _ := collect(t, []string{path}, 0, filter)

	withSinceSearch(t, 256, 512) // small enough that it engages
	skipped, _ := collect(t, []string{path}, 0, filter)

	assert.Equal(t, whole, skipped)
	assert.Len(t, whole, 150)
}

func TestPruneSourcesDropsArchivesOlderThanTheBound(t *testing.T) {
	dir := t.TempDir()
	live := writeAt(t, dir, "node.log", "")
	recent := writeAt(t, dir, "node.archive.log", "")
	ancient := writeAt(t, dir, "node.archive.2.log", "")

	now := time.Now()
	touch(t, recent, now.Add(-time.Hour))
	touch(t, ancient, now.Add(-48*time.Hour))

	sources := []string{live, recent, ancient}
	kept := PruneSources(sources, now.Add(-24*time.Hour))

	assert.Equal(t, []string{live, recent}, kept)
	assert.Equal(t, []string{live, recent, ancient}, sources, "the caller's slice is left alone")
}

// The live log is where the follow offset comes from and the file whose absence
// gets explained, so it stays whatever its timestamp says.
func TestPruneSourcesNeverDropsTheLiveLog(t *testing.T) {
	dir := t.TempDir()
	live := writeAt(t, dir, "node.log", "")
	touch(t, live, time.Now().Add(-48*time.Hour))

	kept := PruneSources([]string{live}, time.Now().Add(-time.Hour))

	assert.Equal(t, []string{live}, kept)
}

func TestPruneSourcesKeepsEverythingWithoutABound(t *testing.T) {
	dir := t.TempDir()
	live := writeAt(t, dir, "node.log", "")
	ancient := writeAt(t, dir, "node.archive.log", "")
	touch(t, ancient, time.Now().Add(-48*time.Hour))

	kept := PruneSources([]string{live, ancient}, time.Time{})

	assert.Equal(t, []string{live, ancient}, kept)
}

// Pruning is what keeps a narrow --since away from a compressed history it has
// no use for. The archive here is not valid gzip, so reading it at all would
// fail the scan.
func TestPrunedArchivesAreNeverOpened(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()
	live := writeAt(t, dir, "node.log", lineAt("warning", "recent", now)+"\n")
	archive := writeAt(t, dir, "node.archive.log.gz", "this is not gzip")
	touch(t, archive, now.Add(-48*time.Hour))

	since := now.Add(-time.Hour)
	got, _ := collect(t, PruneSources([]string{live, archive}, since), 0, Filter{MinLevel: LevelWarn, Since: since})

	assert.Equal(t, []string{"recent"}, got)
}

func touch(t *testing.T, path string, when time.Time) {
	t.Helper()
	require.NoError(t, os.Chtimes(path, when, when))
}
