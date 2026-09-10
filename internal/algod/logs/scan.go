package logs

import (
	"bufio"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"errors"
	"io"
	"io/fs"
	"os"
	"slices"
	"strings"
	"time"
)

// ScanResult reports what a Scan did, beyond the entries it emitted.
type ScanResult struct {
	// Offset is the size of the newest source at the moment it was opened.
	// Passing it to Follow resumes at exactly that point, with no gap and no
	// repeated line.
	Offset int64

	// Count is how many entries were emitted.
	Count int

	// ScanLimited reports that the backward walk stopped at maxScanned before
	// collecting the requested number of entries. It is never set when every
	// match was asked for: that request is served by a forward stream, which
	// reads the whole history however long it is.
	ScanLimited bool

	// Skipped lists the archives that could not be read, and why. Reading the
	// history is best effort behind the live log: an archive is a file the node
	// is in the middle of rotating, renaming and compressing, and losing the
	// whole command to one of those is a far worse answer than losing the
	// archive and saying so.
	Skipped []SkippedSource
}

// SkippedSource is an archive a scan could not read, with the reason it gave.
type SkippedSource struct {
	Path string
	Err  error
}

// errScanBudget ends a forward read of a compressed archive that has used up
// the scan budget. It never reaches the caller; tailCompressed turns it into a
// stop reason.
var errScanBudget = errors.New("scan budget exhausted")

// Scan emits every entry in the log history that passes f, oldest first.
//
// sources are the files that make up that history, newest first: the live
// node.log followed by whatever rotated archives still exist. An archive that
// disappears mid-scan is skipped; the newest source is the live log, and its
// absence is returned so the caller can explain it.
//
// A positive n limits the result to the n newest matches, which is served by
// walking backwards from the end of the newest source and reaching into an
// archive only if the live log did not hold enough. A non-positive n asks for
// everything, which is served by streaming forwards from the oldest source: the
// same bytes are read either way, but nothing is held in memory and the caller
// can print each entry as it arrives rather than waiting out a 1 GiB file.
func Scan(sources []string, n int, f Filter, emit func(Entry) error) (ScanResult, error) {
	var result ScanResult

	if len(sources) == 0 {
		return result, fs.ErrNotExist
	}

	// The live log's size is the follow offset, and it is taken before a byte
	// is read so that nothing written during the scan is missed or repeated.
	info, err := os.Stat(sources[0])
	if err != nil {
		return result, err
	}
	result.Offset = info.Size()

	counted := func(e Entry) error {
		result.Count++
		return emit(e)
	}

	if n > 0 {
		found, err := tailSources(sources, n, f, result.Offset, &result)
		if err != nil {
			return result, err
		}
		for _, e := range found {
			if err := counted(e); err != nil {
				return result, err
			}
		}
		return result, nil
	}

	return result, streamSources(sources, f, result.Offset, counted, &result)
}

// tailSources collects the n newest matching entries across the history,
// oldest first.
//
// maxScanned is one budget for the whole walk rather than one per file: the
// point of the cap is to bound the work a search that matches nothing can do,
// and that is no less true once the search continues into an archive.
func tailSources(sources []string, n int, f Filter, liveSize int64, result *ScanResult) ([]Entry, error) {
	pre := newPrescreen(f)
	budget := maxScanned

	var found []Entry // newest first; reversed before returning

	for i, path := range sources {
		if len(found) >= n {
			break
		}
		if budget <= 0 {
			// Sources are left but there is nothing to spend on them, so this
			// answer is short of what was asked for and has to say so. Without
			// this a budget used up exactly at a file boundary would skip the
			// rest of the history in silence.
			result.ScanLimited = true
			break
		}

		// Only the live log is bounded by the size taken above; an archive is
		// no longer written to, so its whole length is fair game.
		size := int64(-1)
		if i == 0 {
			size = liveSize
		}

		entries, scanned, reason, err := tailOne(path, size, n-len(found), f, pre, budget)
		budget -= scanned
		if err != nil {
			if i == 0 {
				return nil, err // the live log: the caller explains this one
			}
			// Anything at all can be wrong with an archive, and none of it is
			// worth the whole command. It may have been rotated away since it
			// was listed, or still be part way through being compressed, in
			// which case the decompressor stops at an unexpected end of file.
			result.Skipped = append(result.Skipped, SkippedSource{Path: path, Err: err})
			continue
		}

		found = append(found, entries...)

		switch reason {
		case stopBudget:
			// Older entries exist and were not looked at, which the caller
			// reports so the user knows the answer is partial.
			result.ScanLimited = true
		case stopSince:
			// The walk reached the lower time bound. Nothing earlier can match,
			// here or in any archive behind this file, so this is a complete
			// answer rather than a truncated one.
		}
		if reason != stopNone {
			break
		}
	}

	slices.Reverse(found)
	return found, nil
}

// tailOne walks one source backwards for up to need matching entries, choosing
// between the two ways of reaching the end of a file.
func tailOne(path string, size int64, need int, f Filter, pre prescreen, budget int64) ([]Entry, int64, stopReason, error) {
	if IsCompressed(path) {
		return tailCompressed(path, need, f, pre, budget)
	}
	return tailFile(path, size, need, f, pre, budget)
}

// tailCompressed collects the last need matches from a compressed archive.
//
// A gzip or bzip2 stream cannot be read from its end, so the file is read
// forwards and the newest matches are kept in a ring.
//
// The budget is spent on the decompressed bytes, which is the work actually
// being done. Running out of it abandons the archive rather than returning what
// the ring holds: the ring at that point holds the oldest part of the file and
// not its newest, so keeping it would punch a hole in a history that is meant
// to read as one run of entries. What the newer sources gave is a shorter
// answer, which is the honest one, and the caller reports that it is short.
func tailCompressed(path string, need int, f Filter, pre prescreen, budget int64) ([]Entry, int64, stopReason, error) {
	r, err := openSource(path, 0, -1)
	if err != nil {
		return nil, 0, stopNone, err
	}
	defer func() { _ = r.Close() }()

	counter := &countingReader{r: r}

	var (
		ring []Entry // oldest first once full, starting at next
		next int
		full bool
	)
	err = forEachLine(counter, func(line []byte) error {
		if counter.n >= budget {
			return errScanBudget
		}
		e, ok := match(line, f, pre)
		if !ok {
			return nil
		}
		if len(ring) < need {
			ring = append(ring, e)
			return nil
		}
		ring[next] = e
		next = (next + 1) % need
		full = true
		return nil
	})
	if errors.Is(err, errScanBudget) {
		return nil, counter.n, stopBudget, nil
	}
	if err != nil {
		return nil, counter.n, stopNone, err
	}

	if full {
		ring = append(ring[next:], ring[:next]...)
	}
	slices.Reverse(ring) // the caller's convention is newest first
	return ring, counter.n, stopNone, nil
}

// streamSources emits every match in the history, oldest first, holding no
// entries in memory.
func streamSources(sources []string, f Filter, liveSize int64, emit func(Entry) error, result *ScanResult) error {
	pre := newPrescreen(f)

	// Oldest source first, so the output stays chronological across a rotation.
	for i := len(sources) - 1; i >= 0; i-- {
		// Stop the live log where the follow offset was taken. An archive has
		// no such boundary: it is read to its end.
		limit := int64(-1)
		if i == 0 {
			limit = liveSize
		}

		err := streamOne(sources[i], limit, f, pre, emit)
		if err != nil {
			if i == 0 {
				return err // the live log: the caller explains this one
			}
			// Best effort behind the live log, on the same terms as tailSources.
			result.Skipped = append(result.Skipped, SkippedSource{Path: sources[i], Err: err})
			continue
		}
	}
	return nil
}

func streamOne(path string, limit int64, f Filter, pre prescreen, emit func(Entry) error) error {
	offset, err := sinceOffset(path, limit, f.Since)
	if err != nil {
		return err
	}

	r, err := openSource(path, offset, limit)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()

	return forEachLine(r, func(line []byte) error {
		e, ok := match(line, f, pre)
		if !ok {
			return nil
		}
		return emit(e)
	})
}

// probeSize is how much is read for one step of the search in sinceOffset. It
// has to be comfortably larger than a log line, so that a probe landing at an
// arbitrary offset still finds a whole line there to read. It is a var rather
// than a const so that tests can shrink it and exercise the search against a
// small fixture.
var probeSize int64 = 64 << 10

// sinceSlack is how far back the search's answer is rewound before reading
// starts.
//
// Timestamps in node.log are ordered but not perfectly sorted: algod stamps an
// entry when it is created and writes it under a lock, so an entry delayed
// between the two can land behind one that is fractionally newer. The rewind
// makes the answer safe against that, for one extra megabyte of a read that
// would otherwise have been hundreds. A var for the same reason as probeSize.
var sinceSlack int64 = 1 << 20

// sinceOffset returns a byte offset in path at or before the first entry that
// is not older than since, so that a forward read can start there instead of at
// the beginning of the file.
//
// node.log is written in time order, which makes that offset findable by
// bisection: around thirteen probes on a gigabyte, against reading the gigabyte
// to arrive at the same line. It pays off here and only here. A backward walk
// is already travelling towards the bound, and every byte it crosses on the way
// is inside the window it was asked for, so searching for the bound would cost
// it more seeks than the walk saves.
//
// The result is a place to start reading and never a filter. Filter.Keep still
// decides every entry, so an answer that lands too early costs a few pages of
// reading and nothing else. Landing too late would lose entries, which is why
// every step that cannot be decided moves the search downwards.
//
// A compressed archive cannot be seeked into and always gets 0. Those are
// excluded whole, by PruneSources, before a scan ever opens them.
func sinceOffset(path string, limit int64, since time.Time) (int64, error) {
	if since.IsZero() || IsCompressed(path) {
		return 0, nil
	}

	fh, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer func() { _ = fh.Close() }()

	size := limit
	if size < 0 {
		info, err := fh.Stat()
		if err != nil {
			return 0, err
		}
		size = info.Size()
	}

	// Any smaller than the slack the answer gets rewound by and the search
	// cannot save a read worth the seeks it would take.
	if size <= probeSize+sinceSlack {
		return 0, nil
	}

	buf := make([]byte, probeSize)
	var lo int64
	hi := size

	for hi-lo > probeSize {
		mid := lo + (hi-lo)/2

		start, line, ok := probeLine(fh, buf, mid, hi)
		if !ok {
			// No whole line here to judge by, so nothing can be concluded.
			hi = mid
			continue
		}

		ts, ok := peekTime(line)
		if !ok || !ts.Before(since) {
			hi = mid
			continue
		}

		// This line is older than the bound, so every entry in range starts
		// after it.
		lo = start
	}

	lo -= sinceSlack
	if lo <= 0 {
		return 0, nil
	}

	// Reading has to begin at a line boundary, or the first line would arrive
	// as a fragment and be shown as an entry that could not be parsed.
	start, _, ok := probeLine(fh, buf, lo, size)
	if !ok {
		return 0, nil
	}
	return start, nil
}

// probeLine reads the first line beginning at or after at, returning the offset
// it starts at along with its contents. It reads a single buffer, so a line
// longer than that is reported as not found rather than assembled.
func probeLine(fh *os.File, buf []byte, at, limit int64) (int64, []byte, bool) {
	if at >= limit {
		return 0, nil, false
	}

	end := at + int64(len(buf))
	if end > limit {
		end = limit
	}
	n, err := fh.ReadAt(buf[:end-at], at)
	if n == 0 || (err != nil && !errors.Is(err, io.EOF)) {
		return 0, nil, false
	}
	chunk := buf[:n]

	// Anywhere but the start of the file, the bytes up to the first newline are
	// the tail of a line that began earlier and belong to it, not to us.
	var from int
	if at > 0 {
		idx := bytes.IndexByte(chunk, '\n')
		if idx < 0 {
			return 0, nil, false
		}
		from = idx + 1
	}

	rest := chunk[from:]
	idx := bytes.IndexByte(rest, '\n')
	if idx < 0 {
		return 0, nil, false
	}
	return at + int64(from), rest[:idx], true
}

// match applies the prescreen and then the full filter to one raw line.
func match(line []byte, f Filter, pre prescreen) (Entry, bool) {
	if len(bytes.TrimSpace(line)) == 0 || pre.rejects(line) {
		return Entry{}, false
	}
	e := ParseLine(line)
	return e, f.Keep(e)
}

// openSource opens a log file for forward reading, transparently decompressing
// an archive that algod compressed on its way out. A non-negative limit stops
// the read at that many bytes of the file, counted from the start of the file
// and not from offset.
//
// offset skips that many bytes before reading, and is only ever non-zero for an
// uncompressed source: a compressed stream has to be read from its beginning,
// and sinceOffset declines to give an offset for one.
func openSource(path string, offset, limit int64) (io.ReadCloser, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	if offset > 0 {
		if _, err := fh.Seek(offset, io.SeekStart); err != nil {
			_ = fh.Close()
			return nil, err
		}
	}

	var reader io.Reader = fh
	if limit >= 0 {
		remaining := limit - offset
		if remaining < 0 {
			remaining = 0
		}
		reader = io.LimitReader(fh, remaining)
	}

	switch {
	case strings.HasSuffix(path, ".gz"):
		zr, err := gzip.NewReader(reader)
		if err != nil {
			_ = fh.Close()
			return nil, err
		}
		return sourceReader{Reader: zr, closers: []io.Closer{zr, fh}}, nil
	case strings.HasSuffix(path, ".bz2"):
		// compress/bzip2 is decompress-only and has no Close.
		return sourceReader{Reader: bzip2.NewReader(reader), closers: []io.Closer{fh}}, nil
	default:
		return sourceReader{Reader: reader, closers: []io.Closer{fh}}, nil
	}
}

// IsCompressed reports whether a source has to be read from the start because
// it is one of the formats algod hands the rotated log to. It is exported so
// that a caller can tell that a file is no use for the things only an
// uncompressed one supports, following it above all.
func IsCompressed(path string) bool {
	return strings.HasSuffix(path, ".gz") || strings.HasSuffix(path, ".bz2")
}

type sourceReader struct {
	io.Reader
	closers []io.Closer
}

func (s sourceReader) Close() error {
	var err error
	for _, c := range s.closers {
		if cerr := c.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}
	return err
}

// countingReader records how many bytes were read, so that a decompressed
// source can report the work it did against the scan budget.
type countingReader struct {
	r io.Reader
	n int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	return n, err
}

// forEachLine calls fn for each line of r, without the newline.
//
// The slice handed to fn is only valid for the duration of the call. Lines past
// maxLineBytes are truncated to it and their remainder discarded, matching what
// ParseLine would do with them and keeping a runaway line from being buffered
// whole.
func forEachLine(r io.Reader, fn func([]byte) error) error {
	br := bufio.NewReaderSize(r, followBufferSize)

	var carry []byte // set only for a line longer than one buffer
	for {
		chunk, err := br.ReadSlice('\n')

		if errors.Is(err, bufio.ErrBufferFull) {
			if room := maxLineBytes - len(carry); room > 0 {
				if len(chunk) > room {
					chunk = chunk[:room]
				}
				carry = append(carry, chunk...)
			}
			continue
		}

		line := chunk
		if len(carry) > 0 {
			if room := maxLineBytes - len(carry); room > 0 {
				if len(chunk) > room {
					chunk = chunk[:room]
				}
				carry = append(carry, chunk...)
			}
			line = carry
		}

		// A trailing line with no newline is only an entry when the reader
		// simply ran out: a log whose last line is still being written ends
		// that way. Any other error means the bytes are not to be trusted,
		// which is what a half-written compressed archive hands back, and
		// passing that on would show a fragment of the container as an entry.
		if len(line) > 0 && (err == nil || errors.Is(err, io.EOF)) {
			if ferr := fn(line); ferr != nil {
				return ferr
			}
		}
		carry = carry[:0]

		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
	}
}
