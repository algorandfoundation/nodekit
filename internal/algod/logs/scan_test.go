package logs

import (
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withMaxScanned shrinks the backward-walk budget so that a small fixture can
// reach it, the way a real filter that matches nothing reaches it on a 1 GiB log.
func withMaxScanned(t *testing.T, size int64) {
	t.Helper()
	original := maxScanned
	maxScanned = size
	t.Cleanup(func() { maxScanned = original })
}

func writeAt(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func writeGzipped(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	fh, err := os.Create(path)
	require.NoError(t, err)
	zw := gzip.NewWriter(fh)
	_, err = zw.Write([]byte(content))
	require.NoError(t, err)
	require.NoError(t, zw.Close())
	require.NoError(t, fh.Close())
	return path
}

func collect(t *testing.T, sources []string, n int, f Filter) ([]string, ScanResult) {
	t.Helper()
	var got []string
	result, err := Scan(sources, n, f, func(e Entry) error {
		got = append(got, e.Message)
		return nil
	})
	require.NoError(t, err)
	return got, result
}

// The archive holds what the live log used to, so the two read as one history.
func TestScanReadsTheArchiveBehindTheLiveLog(t *testing.T) {
	dir := t.TempDir()
	archive := writeAt(t, dir, "node.archive.log", line("warning", "old one")+"\n"+line("error", "old two")+"\n")
	live := writeAt(t, dir, "node.log", line("warning", "new one")+"\n")

	got, result := collect(t, []string{live, archive}, 0, Filter{MinLevel: LevelWarn})
	assert.Equal(t, []string{"old one", "old two", "new one"}, got)
	assert.Equal(t, 3, result.Count)
	assert.False(t, result.ScanLimited)
}

// A count is filled from the newest end, reaching into the archive only for
// what the live log could not supply.
func TestScanCountReachesIntoTheArchiveOnlyAsNeeded(t *testing.T) {
	dir := t.TempDir()
	archive := writeAt(t, dir, "node.archive.log",
		line("warning", "old one")+"\n"+line("warning", "old two")+"\n"+line("warning", "old three")+"\n")
	live := writeAt(t, dir, "node.log", line("warning", "new one")+"\n")

	got, _ := collect(t, []string{live, archive}, 3, Filter{MinLevel: LevelWarn})
	assert.Equal(t, []string{"old two", "old three", "new one"}, got)

	// Satisfied by the live log alone, the archive is never opened.
	got, _ = collect(t, []string{live, filepath.Join(dir, "absent.log")}, 1, Filter{MinLevel: LevelWarn})
	assert.Equal(t, []string{"new one"}, got)
}

func TestScanSkipsAMissingArchiveButNotAMissingLiveLog(t *testing.T) {
	dir := t.TempDir()
	live := writeAt(t, dir, "node.log", line("warning", "only one")+"\n")
	absent := filepath.Join(dir, "node.archive.log")

	for _, n := range []int{0, 10} {
		got, _ := collect(t, []string{live, absent}, n, Filter{MinLevel: LevelWarn})
		assert.Equal(t, []string{"only one"}, got, "n=%d", n)
	}

	_, err := Scan([]string{absent}, 0, Filter{}, func(Entry) error { return nil })
	assert.True(t, os.IsNotExist(err))

	_, err = Scan(nil, 0, Filter{}, func(Entry) error { return nil })
	assert.Error(t, err)
}

// algod hands the rotated file to gzip when LogArchiveName says so, so the
// archive is often not a plain file.
func TestScanReadsACompressedArchive(t *testing.T) {
	dir := t.TempDir()
	archive := writeGzipped(t, dir, "node.archive.log.gz",
		line("warning", "old one")+"\n"+line("warning", "old two")+"\n"+line("error", "old three")+"\n")
	live := writeAt(t, dir, "node.log", line("warning", "new one")+"\n")

	got, _ := collect(t, []string{live, archive}, 0, Filter{MinLevel: LevelWarn})
	assert.Equal(t, []string{"old one", "old two", "old three", "new one"}, got)

	// A compressed archive cannot be walked from its end, so the newest matches
	// are kept while it is read forwards. The result must not differ.
	got, _ = collect(t, []string{live, archive}, 3, Filter{MinLevel: LevelWarn})
	assert.Equal(t, []string{"old two", "old three", "new one"}, got)

	got, _ = collect(t, []string{live, archive}, 1, Filter{MinLevel: LevelWarn})
	assert.Equal(t, []string{"new one"}, got)
}

// The scan budget exists to stop a search for N entries walking a whole log to
// find them. Asking for every match is not that search: it is meant to reach
// the start, and stopping short would silently drop the oldest entries.
func TestScanBudgetBoundsACountButNotEverything(t *testing.T) {
	var content string
	for i := 0; i < 200; i++ {
		content += line("info", fmt.Sprintf("noise %d", i)) + "\n"
	}
	content = line("warning", "the oldest") + "\n" + content
	path := writeLog(t, content)

	withChunkSize(t, 64)
	withMaxScanned(t, 128)

	got, result := collect(t, []string{path}, 5, Filter{MinLevel: LevelWarn})
	assert.Empty(t, got)
	assert.True(t, result.ScanLimited, "a count must give up at the budget")

	got, result = collect(t, []string{path}, 0, Filter{MinLevel: LevelWarn})
	assert.Equal(t, []string{"the oldest"}, got, "every match means every match")
	assert.False(t, result.ScanLimited)
}

// Offset is the live log's size at the moment the scan started, so Follow
// resumes with no gap and no repeat even though an archive was read too.
func TestScanOffsetIsTheLiveLogSize(t *testing.T) {
	dir := t.TempDir()
	archive := writeAt(t, dir, "node.archive.log", line("warning", "old")+"\n")
	content := line("warning", "new") + "\n"
	live := writeAt(t, dir, "node.log", content)

	_, result := collect(t, []string{live, archive}, 0, Filter{MinLevel: LevelWarn})
	assert.Equal(t, int64(len(content)), result.Offset)
}

// Anything appended after the size was taken belongs to Follow, not to the scan.
func TestScanStopsAtTheOffsetItReported(t *testing.T) {
	dir := t.TempDir()
	live := writeAt(t, dir, "node.log", line("warning", "before")+"\n")

	var got []string
	result, err := Scan([]string{live}, 0, Filter{MinLevel: LevelWarn}, func(e Entry) error {
		got = append(got, e.Message)
		appendLine(t, live, line("warning", "during"))
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"before"}, got)
	assert.Equal(t, 1, result.Count)
}

func TestForEachLineHandlesPartialAndOverlongLines(t *testing.T) {
	// No trailing newline on the last line, and a blank line in between.
	var got []string
	require.NoError(t, forEachLine(strings.NewReader("one\n\ntwo"), func(line []byte) error {
		got = append(got, strings.TrimRight(string(line), "\n"))
		return nil
	}))
	assert.Equal(t, []string{"one", "", "two"}, got)

	// A line past the cap is truncated to it rather than buffered whole, which
	// is what ParseLine would have done with it anyway.
	huge := strings.Repeat("x", maxLineBytes+5000)
	var lengths []int
	require.NoError(t, forEachLine(strings.NewReader(huge+"\nshort\n"), func(line []byte) error {
		lengths = append(lengths, len(strings.TrimRight(string(line), "\n")))
		return nil
	}))
	assert.Equal(t, []int{maxLineBytes, 5}, lengths)
}

func TestScanPropagatesAnEmitError(t *testing.T) {
	path := writeLog(t, line("warning", "one")+"\n"+line("warning", "two")+"\n")
	boom := fmt.Errorf("boom")

	for _, n := range []int{0, 10} {
		_, err := Scan([]string{path}, n, Filter{MinLevel: LevelWarn}, func(Entry) error { return boom })
		assert.ErrorIs(t, err, boom, "n=%d", n)
	}
}

func TestArchiveFilesFindsDatedAndCompressedArchives(t *testing.T) {
	dir := t.TempDir()
	live := writeAt(t, dir, "node.log", "")
	older := writeAt(t, dir, "node.archive-20260908.log.gz", "")
	newer := writeAt(t, dir, "node.archive-20260909.log.gz", "")

	// Names sort in whatever order the template gives, so the order comes from
	// the modification times instead.
	now := time.Now()
	require.NoError(t, os.Chtimes(older, now.Add(-2*time.Hour), now.Add(-2*time.Hour)))
	require.NoError(t, os.Chtimes(newer, now.Add(-1*time.Hour), now.Add(-1*time.Hour)))

	source := Source{
		Path:    live,
		Archive: filepath.Join(dir, "node.archive-{{.Year}}{{.Month}}{{.Day}}.log.gz"),
	}
	assert.Equal(t, []string{newer, older}, source.ArchiveFiles())
	assert.Equal(t, []string{live, newer, older}, source.LogFiles())
}

func TestArchiveFilesIgnoresWhatIsNotThere(t *testing.T) {
	dir := t.TempDir()
	live := writeAt(t, dir, "node.log", "")

	source := Source{Path: live, Archive: filepath.Join(dir, "node.archive.log")}
	assert.Empty(t, source.ArchiveFiles())

	// A directory in the archive's place is not an archive.
	require.NoError(t, os.Mkdir(filepath.Join(dir, "node.archive.log"), 0o755))
	assert.Empty(t, source.ArchiveFiles())

	// Nor is the live log itself, however the config points at it.
	assert.Empty(t, Source{Path: live, Archive: live}.ArchiveFiles())
	assert.Empty(t, Source{Path: live}.ArchiveFiles())
}
