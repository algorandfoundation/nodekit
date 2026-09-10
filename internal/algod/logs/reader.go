package logs

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"
)

// chunkSize is how much is read per step when walking backwards from the end of
// the log. It is a var rather than a const so that tests can shrink it and
// exercise the chunk-boundary handling against small fixtures.
var chunkSize int64 = 64 * 1024

// maxScanned bounds how far back a backward walk will read before giving up. A
// busy node.log routinely reaches algod's 1 GiB LogSizeLimit, and a filter that
// matches nothing would otherwise walk the whole file.
var maxScanned int64 = 512 << 20

// ScanLimit is how far back a request for a fixed number of entries will read
// before giving up, so that the command can say how much it searched.
func ScanLimit() int64 { return maxScanned }

// followBufferSize is the read buffer used while streaming.
const followBufferSize = 64 * 1024

// defaultFollowInterval is how often Follow polls for growth, rotation and
// truncation. Polling is used rather than fsnotify: a rename does raise an
// event, but truncation in place does not, and catching that needs the size
// compared against the read offset on a timer regardless.
const defaultFollowInterval = 250 * time.Millisecond

// stopReason says why a backward walk ended before reaching the start of its
// range. The distinction matters to the caller: only one of these means there
// is more that was not looked at.
type stopReason int

const (
	// stopNone: the range was read to its end, or enough entries were found.
	stopNone stopReason = iota

	// stopBudget: the byte budget ran out with entries still to look for. Older
	// entries exist and were not read, which is worth telling the user.
	stopBudget

	// stopSince: a line older than the filter's lower time bound was reached.
	// Everything earlier is out of range by definition, in this file and in
	// every archive behind it, so stopping here has cost the caller nothing.
	stopSince
)

// tailFile walks the file at path backwards from size, collecting up to need
// matching entries, newest first. A negative size means the whole file.
//
// The file is walked backwards in chunks rather than scanned from the start:
// node.log is commonly hundreds of megabytes, and the entries anyone wants are
// almost always near the end. budget bounds how far back it will read; the walk
// stops there and says so, and it returns what it did read so that a caller
// covering several files can share one budget.
func tailFile(path string, size int64, need int, f Filter, pre prescreen, budget int64) ([]Entry, int64, stopReason, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, 0, stopNone, err
	}
	defer func() { _ = fh.Close() }()

	if size < 0 {
		info, err := fh.Stat()
		if err != nil {
			return nil, 0, stopNone, err
		}
		size = info.Size()
	}

	var (
		found   []Entry // newest first
		partial []byte  // start of a line whose remainder lies in an earlier chunk
		scanned int64
		pos     = size
	)

	for pos > 0 {
		if len(found) >= need {
			break
		}
		if scanned >= budget {
			return found, scanned, stopBudget, nil
		}

		step := chunkSize
		if pos < step {
			step = pos
		}
		pos -= step

		buf := make([]byte, step)
		if _, err := fh.ReadAt(buf, pos); err != nil && err != io.EOF {
			return nil, scanned, stopNone, err
		}
		scanned += step

		if len(partial) > 0 {
			buf = append(buf, partial...)
			partial = nil
		}

		segment := buf
		if pos > 0 {
			// The first line in this chunk began in an earlier one; carry it.
			idx := bytes.IndexByte(buf, '\n')
			if idx < 0 {
				if int64(len(buf)) > maxLineBytes {
					// A single line longer than the cap: stop growing the carry
					// buffer, the excess would be truncated by ParseLine anyway.
					partial = buf[:maxLineBytes]
				} else {
					partial = buf
				}
				continue
			}
			partial = buf[:idx+1]
			segment = buf[idx+1:]
		}

		lines := linesReverse(segment)
		for _, line := range lines {
			e, ok := match(line, f, pre)
			if !ok {
				continue
			}
			found = append(found, e)
			if len(found) >= need {
				break
			}
		}

		// Stop once this chunk has been walked past the filter's lower time
		// bound. Without this the walk runs to the budget looking for entries
		// that cannot exist, and then reports having stopped early, which sends
		// the user off to rerun the same search over even more of the file.
		//
		// The check is per chunk and not per line on purpose. The prescreen
		// exists to keep ParseLine off the great majority of lines a scan
		// touches, and reading a timestamp from each of them would hand that
		// saving straight back. Lines inside a chunk are contiguous in time, so
		// the oldest one it holds is enough to decide, at one read per chunk.
		//
		// The whole chunk is walked before stopping, because the chunk that
		// crosses the bound still holds entries above it. A chunk whose oldest
		// line carries no readable timestamp decides nothing and the walk
		// continues, which costs a chunk and stays correct.
		if len(found) < need && !f.Since.IsZero() && len(lines) > 0 {
			// linesReverse yields newest first, so the last is the oldest line
			// this chunk holds in full.
			if ts, ok := peekTime(lines[len(lines)-1]); ok && ts.Before(f.Since) {
				return found, scanned, stopSince, nil
			}
		}
	}

	// No leftover handling is needed here: partial is folded into the next
	// chunk at the top of each iteration, and the final iteration (pos == 0)
	// consumes it, so it is only ever non-empty when the walk stopped early
	// with a genuinely incomplete line.

	return found, scanned, stopNone, nil
}

// linesReverse splits a buffer into whole lines, newest last line first. A
// single trailing newline is not treated as starting an extra empty line.
func linesReverse(segment []byte) [][]byte {
	if len(segment) == 0 {
		return nil
	}
	if segment[len(segment)-1] == '\n' {
		segment = segment[:len(segment)-1]
	}

	var out [][]byte
	end := len(segment)
	for {
		idx := bytes.LastIndexByte(segment[:end], '\n')
		out = append(out, segment[idx+1:end])
		if idx < 0 {
			break
		}
		end = idx
	}
	return out
}

// FollowOptions tunes Follow.
type FollowOptions struct {
	// Interval is the poll period; defaults to defaultFollowInterval.
	Interval time.Duration

	// OnRotate, when set, is called after the log is rotated or truncated
	// underneath us and reading has restarted on the new file.
	OnRotate func()
}

// Follow streams entries appended to the log at path from offset onwards,
// calling emit for each one that passes the filter, until ctx is cancelled.
//
// Cancellation is not an error: Follow returns nil. It handles the log being
// rotated to its archive or truncated in place while streaming.
func Follow(ctx context.Context, path string, offset int64, f Filter, emit func(Entry) error, opts FollowOptions) error {
	// A compressed archive is not a file anything is appending to, and an
	// offset into one counts decompressed bytes, which means nothing to a seek.
	// Streaming it would read the container itself as though it were text.
	if IsCompressed(path) {
		return fmt.Errorf("cannot follow %s: a compressed archive is not being written to", path)
	}

	interval := opts.Interval
	if interval <= 0 {
		interval = defaultFollowInterval
	}

	fh, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = fh.Close() }()

	if _, err = fh.Seek(offset, io.SeekStart); err != nil {
		return err
	}
	info, err := fh.Stat()
	if err != nil {
		return err
	}

	pre := newPrescreen(f)
	reader := bufio.NewReaderSize(fh, followBufferSize)
	var pending []byte

	// drain consumes every complete line currently available, holding back a
	// trailing partial write until its newline arrives.
	drain := func() error {
		for {
			chunk, err := reader.ReadBytes('\n')
			if err == io.EOF {
				// An unterminated tail is a write caught in flight, not an
				// entry. Hold it until its newline arrives, bounded so that a
				// runaway line cannot grow the buffer without limit.
				if room := maxLineBytes - len(pending); room > 0 && len(chunk) > 0 {
					if len(chunk) > room {
						chunk = chunk[:room]
					}
					pending = append(pending, chunk...)
				}
				return nil
			}
			if err != nil {
				return err
			}

			line := chunk
			if len(pending) > 0 {
				line = append(pending, chunk...)
				pending = nil
			}
			e, ok := match(line, f, pre)
			if !ok {
				continue
			}
			if err := emit(e); err != nil {
				return err
			}
		}
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		// Draining before the first wait flushes anything written between the
		// offset the scan reported and now.
		if err := drain(); err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}

		current, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				// Momentarily absent mid-rotation; keep polling.
				continue
			}
			return err
		}

		if !os.SameFile(info, current) {
			// Rotated. Take whatever is left in the old file before switching.
			if err := drain(); err != nil {
				return err
			}
			replacement, err := os.Open(path)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return err
			}
			replacementInfo, err := replacement.Stat()
			if err != nil {
				_ = replacement.Close()
				return err
			}
			_ = fh.Close()
			fh, info = replacement, replacementInfo
			reader.Reset(fh)
			pending = nil
			if opts.OnRotate != nil {
				opts.OnRotate()
			}
			continue
		}

		// The buffer is empty straight after a drain, so the file offset is the
		// true read position and shrinkage means truncation in place.
		where, err := fh.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}
		if current.Size() < where {
			if _, err = fh.Seek(0, io.SeekStart); err != nil {
				return err
			}
			reader.Reset(fh)
			pending = nil
			if opts.OnRotate != nil {
				opts.OnRotate()
			}
		}
	}
}
