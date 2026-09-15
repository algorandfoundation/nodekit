package logs

// tailResult is the outcome of a tailFilter call.
type tailResult struct {
	// Entries are the matching entries, oldest first.
	Entries []Entry

	// Offset is the file size observed at the moment of reading.
	Offset int64

	// ScanLimited reports that the backward walk stopped at maxScanned before
	// collecting the requested number of entries.
	ScanLimited bool
}

// tailFilter collects up to n matching entries from the end of a single file,
// oldest first. A non-positive n collects every match.
//
// This is the buffered single-file shape the reader had before Scan covered the
// archives and emitted as it went. Nothing outside the tests wants it: the
// command streams instead, so that a gigabyte of log is not held in memory.
// It stays here because most of the tests below are about one file's contents
// and read better against a slice than against a callback.
func tailFilter(path string, n int, f Filter) (tailResult, error) {
	var out tailResult

	result, err := Scan([]string{path}, n, f, func(e Entry) error {
		out.Entries = append(out.Entries, e)
		return nil
	})
	out.Offset, out.ScanLimited = result.Offset, result.ScanLimited
	return out, err
}
