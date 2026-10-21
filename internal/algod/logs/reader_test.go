package logs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withChunkSize shrinks the backward-read step so that small fixtures still
// exercise the chunk-boundary handling that only shows up on real logs.
func withChunkSize(t *testing.T, size int64) {
	t.Helper()
	original := chunkSize
	chunkSize = size
	t.Cleanup(func() { chunkSize = original })
}

func writeLog(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "node.log")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

// openLog opens a fixture the way Scan opens a source, for the functions that
// read one through a handle.
func openLog(t *testing.T, path string) *os.File {
	t.Helper()
	fh, err := os.Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = fh.Close() })
	return fh
}

func line(level, msg string) string {
	return fmt.Sprintf(`{"level":%q,"msg":%q,"time":"2026-09-09T14:16:19+02:00"}`, level, msg)
}

func messages(entries []Entry) []string {
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.Message)
	}
	return out
}

func TestTailFilterReturnsOldestFirst(t *testing.T) {
	path := writeLog(t, line("warning", "one")+"\n"+line("warning", "two")+"\n"+line("warning", "three")+"\n")

	result, err := tailFilter(path, 2, Filter{MinLevel: LevelWarn})
	require.NoError(t, err)
	assert.Equal(t, []string{"two", "three"}, messages(result.Entries))
}

// The count applies to entries that matched, not to lines of the file. With
// warnings at well under one percent of a real log, counting raw lines would
// make the default view almost always empty.
func TestTailFilterCountsMatchesNotLines(t *testing.T) {
	var content string
	for i := 0; i < 500; i++ {
		content += line("info", fmt.Sprintf("noise %d", i)) + "\n"
	}
	content += line("warning", "needle") + "\n"
	for i := 0; i < 500; i++ {
		content += line("info", fmt.Sprintf("more noise %d", i)) + "\n"
	}

	withChunkSize(t, 64)
	result, err := tailFilter(writeLog(t, content), 10, Filter{MinLevel: LevelWarn})
	require.NoError(t, err)
	assert.Equal(t, []string{"needle"}, messages(result.Entries))
}

func TestTailFilterAcrossChunkBoundaries(t *testing.T) {
	var content string
	for i := 0; i < 40; i++ {
		content += line("warning", fmt.Sprintf("entry %d", i)) + "\n"
	}

	// A chunk far smaller than one line forces every line to be reassembled
	// from several chunks. Only a count walks backwards, so the boundary
	// handling is exercised with one; n of 0 is checked alongside it because
	// the two paths have to agree.
	for _, size := range []int64{1, 7, 64, 1024} {
		for _, n := range []int{0, 40, 100} {
			t.Run(fmt.Sprintf("chunk-%d-n-%d", size, n), func(t *testing.T) {
				withChunkSize(t, size)
				result, err := tailFilter(writeLog(t, content), n, Filter{MinLevel: LevelWarn})
				require.NoError(t, err)
				require.Len(t, result.Entries, 40)
				assert.Equal(t, "entry 0", result.Entries[0].Message)
				assert.Equal(t, "entry 39", result.Entries[39].Message)
			})
		}
	}
}

func TestTailFilterWithoutTrailingNewline(t *testing.T) {
	withChunkSize(t, 16)
	path := writeLog(t, line("warning", "first")+"\n"+line("warning", "last"))

	for _, n := range []int{0, 5} {
		result, err := tailFilter(path, n, Filter{MinLevel: LevelWarn})
		require.NoError(t, err)
		assert.Equal(t, []string{"first", "last"}, messages(result.Entries), "n=%d", n)
	}
}

// The only match being the very first line forces the walk all the way back to
// the start of the file.
func TestTailFilterFindsFirstLine(t *testing.T) {
	content := line("error", "the only one") + "\n"
	for i := 0; i < 200; i++ {
		content += line("info", fmt.Sprintf("noise %d", i)) + "\n"
	}

	withChunkSize(t, 32)
	result, err := tailFilter(writeLog(t, content), 5, Filter{MinLevel: LevelWarn})
	require.NoError(t, err)
	assert.Equal(t, []string{"the only one"}, messages(result.Entries))
}

func TestTailFilterEmptyFile(t *testing.T) {
	result, err := tailFilter(writeLog(t, ""), 10, Filter{MinLevel: LevelWarn})
	require.NoError(t, err)
	assert.Empty(t, result.Entries)
	assert.Zero(t, result.Offset)
}

func TestTailFilterBlankLinesAreNotEntries(t *testing.T) {
	path := writeLog(t, "\n\n"+line("warning", "real")+"\n\n\n")
	result, err := tailFilter(path, 0, Filter{MinLevel: LevelWarn})
	require.NoError(t, err)
	assert.Equal(t, []string{"real"}, messages(result.Entries))
}

func TestTailFilterKeepsPlainLines(t *testing.T) {
	path := writeLog(t, line("info", "starting")+"\npanic: runtime error\n")
	result, err := tailFilter(path, 0, Filter{MinLevel: LevelWarn})
	require.NoError(t, err)
	assert.Equal(t, []string{"panic: runtime error"}, messages(result.Entries))
}

// Offset is the size at the moment of reading, which is what lets Follow resume
// without a gap or a repeated line.
func TestTailFilterOffsetIsFileSize(t *testing.T) {
	content := line("warning", "one") + "\n"
	path := writeLog(t, content)

	result, err := tailFilter(path, 0, Filter{MinLevel: LevelWarn})
	require.NoError(t, err)
	assert.Equal(t, int64(len(content)), result.Offset)
}

// A stale size, taken before the file was truncated or replaced by a shorter
// one, leaves the tail of the read buffer zeroed. Those bytes are not log
// content: charging them invents entries out of the padding and spends the
// scan budget on a region of the file that no longer exists.
func TestTailFileWithAStaleSize(t *testing.T) {
	withChunkSize(t, 128)

	path := writeLog(t, line("warning", "one")+"\n"+line("warning", "two")+"\n")
	info, err := os.Stat(path)
	require.NoError(t, err)

	f := Filter{MinLevel: LevelTrace}
	found, _, _, err := tailFile(openLog(t, path), info.Size()+3*chunkSize, 10, f, newPrescreen(f), maxScanned)
	require.NoError(t, err)
	assert.Equal(t, []string{"two", "one"}, messages(found)) // tailFile yields newest first
}

func TestTailFilterMissingFile(t *testing.T) {
	_, err := tailFilter(filepath.Join(t.TempDir(), "absent.log"), 10, Filter{})
	assert.True(t, os.IsNotExist(err))
}

// collector gathers followed entries in a goroutine-safe way.
type collector struct {
	mu      sync.Mutex
	entries []Entry
}

func (c *collector) emit(e Entry) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = append(c.entries, e)
	return nil
}

func (c *collector) messages() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return messages(c.entries)
}

func (c *collector) waitFor(t *testing.T, count int) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		c.mu.Lock()
		got := len(c.entries)
		c.mu.Unlock()
		if got >= count {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("timed out waiting for %d entries, got %v", count, c.messages())
}

func appendLine(t *testing.T, path, content string) {
	t.Helper()
	fh, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o600)
	require.NoError(t, err)
	_, err = fh.WriteString(content)
	require.NoError(t, err)
	require.NoError(t, fh.Close())
}

func startFollow(t *testing.T, path string, offset int64, c *collector, opts FollowOptions) context.CancelFunc {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- Follow(ctx, path, offset, Filter{MinLevel: LevelTrace}, c.emit, opts)
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case err := <-done:
			assert.NoError(t, err)
		case <-time.After(3 * time.Second):
			t.Error("Follow did not return after cancellation")
		}
	})
	return cancel
}

func TestFollowStreamsAppendedLines(t *testing.T) {
	path := writeLog(t, line("info", "existing")+"\n")
	info, err := os.Stat(path)
	require.NoError(t, err)

	var c collector
	startFollow(t, path, info.Size(), &c, FollowOptions{Interval: time.Millisecond})

	appendLine(t, path, line("warning", "appended")+"\n")
	c.waitFor(t, 1)
	assert.Equal(t, []string{"appended"}, c.messages())
}

// A line only becomes an entry once its newline arrives, otherwise a log write
// caught mid-flight would be reported as a truncated entry.
func TestFollowHoldsPartialLines(t *testing.T) {
	path := writeLog(t, "")
	var c collector
	startFollow(t, path, 0, &c, FollowOptions{Interval: time.Millisecond})

	complete := line("warning", "split")
	appendLine(t, path, complete[:20])
	time.Sleep(20 * time.Millisecond)
	assert.Empty(t, c.messages(), "a partial line must not be emitted")

	appendLine(t, path, complete[20:]+"\n")
	c.waitFor(t, 1)
	assert.Equal(t, []string{"split"}, c.messages())
}

// A single line longer than the read buffer must not be buffered whole: it is
// capped where ParseLine would have capped it, and the discarded remainder does
// not disturb the lines after it.
func TestFollowCapsARunawayLine(t *testing.T) {
	path := writeLog(t, "")
	var c collector
	startFollow(t, path, 0, &c, FollowOptions{Interval: time.Millisecond})

	appendLine(t, path, strings.Repeat("x", maxLineBytes+5000)+"\n")
	appendLine(t, path, line("warning", "after")+"\n")
	c.waitFor(t, 2)

	c.mu.Lock()
	defer c.mu.Unlock()
	assert.Equal(t, maxLineBytes, len(c.entries[0].Raw))
	assert.Equal(t, "after", c.entries[1].Message)
}

// algod rotates node.log to its archive once LogSizeLimit is reached. Anything
// already written to the old file must still be delivered.
func TestFollowSurvivesRotation(t *testing.T) {
	path := writeLog(t, "")
	var c collector
	rotated := make(chan struct{}, 1)
	startFollow(t, path, 0, &c, FollowOptions{
		Interval: time.Millisecond,
		OnRotate: func() {
			select {
			case rotated <- struct{}{}:
			default:
			}
		},
	})

	appendLine(t, path, line("warning", "before rotation")+"\n")
	c.waitFor(t, 1)

	require.NoError(t, os.Rename(path, path+".archive"))
	require.NoError(t, os.WriteFile(path, []byte(line("warning", "after rotation")+"\n"), 0o600))

	c.waitFor(t, 2)
	assert.Equal(t, []string{"before rotation", "after rotation"}, c.messages())

	select {
	case <-rotated:
	case <-time.After(time.Second):
		t.Error("OnRotate was never called")
	}
}

func TestFollowSurvivesTruncation(t *testing.T) {
	path := writeLog(t, "")
	var c collector
	startFollow(t, path, 0, &c, FollowOptions{Interval: time.Millisecond})

	appendLine(t, path, line("warning", "before truncation")+"\n")
	c.waitFor(t, 1)

	require.NoError(t, os.Truncate(path, 0))
	// The poller sees a truncation by the file being shorter than the offset it
	// had read to, so it has to look while the file is still short. Growing it
	// back past that offset first would hide the truncation entirely -- a known
	// and deliberate limit of polling, recorded at the check in Follow -- and
	// this test is about the detection that does work, not that window.
	time.Sleep(20 * time.Millisecond)
	appendLine(t, path, line("warning", "after truncation")+"\n")

	c.waitFor(t, 2)
	assert.Equal(t, []string{"before truncation", "after truncation"}, c.messages())
}

// The log can rotate or be truncated between the scan that reported an offset
// and Follow opening the file. The offset then points past the end of a file it
// was never measured against, and seeking there would skip everything the
// replacement already holds and wait for it to grow back to a position that
// means nothing in it.
//
// The poll interval is long enough never to fire: the offset has to be settled
// when the file is opened, and not by the truncation check a tick later, which
// only catches this if the new file has not already grown past the stale offset.
func TestFollowResumesFromTheStartOfAShorterFile(t *testing.T) {
	path := writeLog(t, line("warning", "written after the rotation")+"\n")

	var c collector
	rotated := make(chan struct{}, 1)
	startFollow(t, path, 1<<20, &c, FollowOptions{
		Interval: time.Hour,
		OnRotate: func() {
			select {
			case rotated <- struct{}{}:
			default:
			}
		},
	})

	c.waitFor(t, 1)
	assert.Equal(t, []string{"written after the rotation"}, c.messages())

	select {
	case <-rotated:
	case <-time.After(time.Second):
		t.Error("restarting at the top of a new file is a rotation and must be reported as one")
	}
}

// rotate does what algod does once LogSizeLimit is reached: renames the live
// log to its archive and starts a new one, which here already holds lines.
func rotate(t *testing.T, path string, lines []string) {
	t.Helper()
	require.NoError(t, os.Rename(path, path+".archive"))
	require.NoError(t, os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600))
}

// The log can rotate between the scan that reports an offset and Follow picking
// up from it, and with --follow replaying the whole history by default that gap
// lasts as long as the replay. Whatever was written to the old file in the gap
// has to arrive, and then the new file from its start. Reopening the path loses
// the first, and loses part of the second too when the new file has already
// grown past the offset, because its size then gives no sign of the swap.
func TestFollowReadsTheScannedFileToItsEndAcrossARotation(t *testing.T) {
	for _, tc := range []struct {
		name        string
		replacement int
	}{
		{"replaced by a file shorter than the offset", 1},
		{"replaced by a file longer than the offset", 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			scanned := []string{line("warning", "scanned 1"), line("warning", "scanned 2"), line("warning", "scanned 3")}
			path := writeLog(t, strings.Join(scanned, "\n")+"\n")

			result, err := Scan([]string{path}, 0, Filter{MinLevel: LevelTrace}, func(Entry) error { return nil })
			require.NoError(t, err)
			t.Cleanup(func() { _ = result.Close() })

			appendLine(t, path, line("warning", "written during the scan")+"\n")

			want := []string{"written during the scan"}
			var replacement []string
			for i := range tc.replacement {
				msg := fmt.Sprintf("new-file-%d", i)
				replacement = append(replacement, line("warning", msg))
				want = append(want, msg)
			}
			rotate(t, path, replacement)

			var c collector
			rotated := make(chan struct{}, 1)
			startFollow(t, path, result.Offset, &c, FollowOptions{
				Interval: time.Millisecond,
				Live:     result.Live,
				OnRotate: func() {
					select {
					case rotated <- struct{}{}:
					default:
					}
				},
			})

			c.waitFor(t, len(want))
			assert.Equal(t, want, c.messages())

			select {
			case <-rotated:
			case <-time.After(time.Second):
				t.Error("moving on to the new file is a rotation and must be reported as one")
			}
		})
	}
}

// A rotation part way through the scan itself, while an archive is being
// replayed, must not swap the live log the scan goes on to read for the file
// that replaced it. The scan shows the file it measured, and Follow carries on
// from there: the old file's tail, then the new file.
func TestScanAndFollowAcrossARotationDuringTheScan(t *testing.T) {
	dir := t.TempDir()
	archive := writeAt(t, dir, "node.log.archive", line("warning", "archived")+"\n")
	path := writeAt(t, dir, "node.log", line("warning", "live")+"\n")

	// The archive is named the way rotate names it, so the rotation renames the
	// live log onto the archive while the scan is reading it. That is what
	// algod does, reusing one archive name, and it leaves the archive's handle
	// on the file it was listed as, so the scan still reads it whole.
	var scanned []string
	result, err := Scan([]string{path, archive}, 0, Filter{MinLevel: LevelTrace}, func(e Entry) error {
		if e.Message == "archived" {
			appendLine(t, path, line("warning", "written during the scan")+"\n")
			rotate(t, path, []string{line("warning", "new file")})
		}
		scanned = append(scanned, e.Message)
		return nil
	})
	require.NoError(t, err)
	t.Cleanup(func() { _ = result.Close() })
	assert.Equal(t, []string{"archived", "live"}, scanned, "the live log as it was measured, not its replacement")

	var c collector
	rotated := make(chan struct{}, 1)
	startFollow(t, path, result.Offset, &c, FollowOptions{
		Interval: time.Millisecond,
		Live:     result.Live,
		OnRotate: func() {
			select {
			case rotated <- struct{}{}:
			default:
			}
		},
	})

	c.waitFor(t, 2)
	assert.Equal(t, []string{"written during the scan", "new file"}, c.messages())

	select {
	case <-rotated:
	case <-time.After(time.Second):
		t.Error("moving on to the new file is a rotation and must be reported as one")
	}
}

func TestFollowWaitsForALogThatIsMidRotation(t *testing.T) {
	// The live path is absent when Follow opens it, the way it is between the
	// rename and the create of a rotation. The scan has already read the file
	// by this point, so this is a rename in flight, not a node that never ran.
	path := filepath.Join(t.TempDir(), "node.log")

	var c collector
	startFollow(t, path, 0, &c, FollowOptions{Interval: time.Millisecond})
	time.Sleep(20 * time.Millisecond)

	require.NoError(t, os.WriteFile(path, []byte(line("warning", "the replacement file")+"\n"), 0o600))

	c.waitFor(t, 1)
	assert.Equal(t, []string{"the replacement file"}, c.messages())
}

func TestFollowFailsFastOnAnErrorThatWillNotResolve(t *testing.T) {
	// A path whose parent is a regular file gives ENOTDIR, standing in for the
	// class of error that waiting cannot fix. Only absence is worth waiting
	// out; a permission denied asked again is a permission denied.
	path := filepath.Join(writeLog(t, ""), "node.log")

	done := make(chan error, 1)
	go func() {
		done <- Follow(context.Background(), path, 0, Filter{}, func(Entry) error { return nil }, FollowOptions{Interval: time.Hour})
	}()

	select {
	case err := <-done:
		require.Error(t, err)
		assert.False(t, os.IsNotExist(err), "the error waited on must be absence and nothing else")
	case <-time.After(3 * time.Second):
		t.Fatal("Follow waited on an error that was never going to resolve")
	}
}

func TestFollowReturnsNilOnCancel(t *testing.T) {
	path := writeLog(t, "")
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- Follow(ctx, path, 0, Filter{}, func(Entry) error { return nil }, FollowOptions{Interval: time.Millisecond})
	}()

	cancel()
	select {
	case err := <-done:
		assert.NoError(t, err, "cancellation is how the user stops following, not an error")
	case <-time.After(3 * time.Second):
		t.Fatal("Follow did not return after cancellation")
	}
}
