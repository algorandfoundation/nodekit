package logs

import (
	"bytes"
	"encoding/json"
	"time"
)

// levelKey is how logrus writes the severity field into node.log. The value
// follows immediately, so finding this is enough to read the level without
// decoding the line.
var (
	levelKey  = []byte(`"level":"`)
	levelName = []byte("level")
	timeKey   = []byte(`"time":"`)
	timeName  = []byte("time")
)

// prescreen is the part of a Filter that can be decided from the raw bytes of a
// line, before paying for ParseLine.
//
// This matters because ParseLine would otherwise run on every line a scan
// touches, including the ones about to be discarded: it copies the line, decodes it into
// a map and builds a json.Decoder per field, which measures at roughly 3.7µs
// and 43 allocations per line. At the default warn floor that work is wasted on
// the ~99% of a node.log that is info. Deciding those from the raw bytes costs
// about 370ns and no allocation, which BenchmarkFilterWithPrescreen puts at
// 10x the throughput of the same scan without it, for an eightieth of the
// garbage. A --filter does better still, since its text can be looked for in
// the raw line with a substring search alone.
//
// A prescreen is one-sided: it only ever rejects lines that the full Filter
// would also reject. Anything it cannot decide falls through to ParseLine, so
// unparseable lines, missing levels and unrecognised levels keep their
// LevelUnknown handling in Filter.Keep.
type prescreen struct {
	minLevel Level
	textLit  []byte
}

// newPrescreen derives the cheap checks from f once, outside the scan loop.
func newPrescreen(f Filter) prescreen {
	return prescreen{minLevel: f.MinLevel, textLit: searchable(f.Text)}
}

// rejects reports whether line can be discarded without parsing it.
func (p prescreen) rejects(line []byte) bool {
	// The search text goes first: it is the cheaper of the two checks and,
	// when the user supplied one, by far the more selective.
	//
	// The decoded message is a byte-for-byte substring of the raw line whenever
	// the text needs no JSON escaping, so its absence from the line means it
	// cannot match the message. Note the converse does not hold: the text may
	// appear in a field the filter does not look at, such as "function", which
	// is why this only prefilters and Filter.Keep still runs on the entry.
	if p.textLit != nil && !bytes.Contains(line, p.textLit) {
		return true
	}

	// At the lowest floor no level can fall below it, so there is nothing to
	// gain from looking.
	if p.minLevel > LevelTrace {
		if level, ok := peekLevel(line); ok && level < p.minLevel {
			return true
		}
	}

	return false
}

// peekLevel reads the severity out of a raw line without decoding it into an
// Entry. The second return is false when no level could be read with certainty,
// in which case the caller must draw no conclusion and parse the line properly.
//
// Certainty is the whole difficulty here, because the line this says "info"
// about is a line that will never be shown. ParseLine gives an unreadable line
// LevelUnknown, which Filter.Keep then keeps at any floor up to error, so
// claiming a level for a line ParseLine would reject silently loses it. Both of
// the shortcuts worth trying are wrong, and FuzzPrescreenNeverRejectsAKeptLine
// finds each within a minute:
//
//   - Taking the bytes after the first "level":" reads a nested level as the
//     entry's own. Sorted keys put "details" ahead of "level", and algod writes
//     telemetry entries with objects under it.
//   - Requiring that key to sit after a { or a , does not fix it, because a
//     nested key sits there too: {"":{"level":"info"}} has no level of its own.
//
// So the level is located at object depth one, and the line is validated before
// its level is believed. json.Valid costs around 155ns against ParseLine's
// 3.8µs, which leaves the prefilter well worth having.
func peekLevel(line []byte) (Level, bool) {
	line = bytes.TrimSpace(line)

	// Longer lines are truncated by ParseLine, and what remains no longer
	// parses, so no level read from one here would survive.
	if len(line) == 0 || line[0] != '{' || len(line) > maxLineBytes {
		return LevelUnknown, false
	}

	// A cheap negative first: without the key anywhere there is nothing to find,
	// and this costs a fraction of the scan below.
	if !bytes.Contains(line, levelKey) {
		return LevelUnknown, false
	}

	value, ok := topLevelString(line, levelName)
	if !ok {
		return LevelUnknown, false
	}

	// The spellings logrus actually writes. "warning" is the one that matters:
	// matching only "warn" would quietly prefilter away every warning.
	var level Level
	switch string(value) {
	case "trace":
		level = LevelTrace
	case "debug":
		level = LevelDebug
	case "info":
		level = LevelInfo
	case "warning", "warn":
		level = LevelWarn
	case "error":
		level = LevelError
	case "fatal":
		level = LevelFatal
	case "panic":
		level = LevelPanic
	default:
		// An unrecognised level is LevelUnknown, which Filter.Keep deliberately
		// keeps, so report that we could not decide rather than that it is low.
		return LevelUnknown, false
	}

	if !json.Valid(line) {
		return LevelUnknown, false
	}
	return level, true
}

// peekTime reads the timestamp out of a raw line on the same terms as
// peekLevel, without decoding it into an Entry. A false second return means
// nothing could be read with certainty, and the caller must draw no conclusion
// from it.
//
// This exists so that a scan can tell how far through the log it has walked
// without parsing every line it passes. Calling it per line would give back
// what the prescreen is for, so callers use it per chunk of the file instead.
func peekTime(line []byte) (time.Time, bool) {
	line = bytes.TrimSpace(line)

	if len(line) == 0 || line[0] != '{' || len(line) > maxLineBytes {
		return time.Time{}, false
	}

	if !bytes.Contains(line, timeKey) {
		return time.Time{}, false
	}

	value, ok := topLevelString(line, timeName)
	if !ok {
		return time.Time{}, false
	}

	// topLevelString hands back the bytes as written. A timestamp algod wrote
	// carries no escape, so anything that fails to parse here is a line whose
	// time we have no business believing.
	ts, err := time.Parse(time.RFC3339Nano, string(value))
	if err != nil {
		return time.Time{}, false
	}

	// topLevelString relies on a key being what a colon follows, which only
	// holds for a well-formed line. peekLevel validates for the same reason.
	if !json.Valid(line) {
		return time.Time{}, false
	}
	return ts, true
}

// topLevelString returns the string value held by key in the outermost object
// of line, and whether exactly one such key was found holding a plain string.
// The bytes come back exactly as written, escapes and all, which is enough for
// the caller: none of the level names contains one.
//
// It walks the line once, tracking nesting so that a key of the same name
// inside a nested object or array is not mistaken for the outer one, and
// skipping over string contents so that a key spelled out inside a message is
// not mistaken for a key. Callers must confirm the line is valid JSON: the
// distinction between a key and a string value relies on a key being the thing
// followed by a colon.
func topLevelString(line, key []byte) ([]byte, bool) {
	var (
		found []byte
		seen  bool
		depth int
	)

	for i := 0; i < len(line); {
		switch line[i] {
		case '{', '[':
			depth++
			i++
		case '}', ']':
			depth--
			i++
		case '"':
			end := i + 1
			for end < len(line) && line[end] != '"' {
				if line[end] == '\\' {
					end++
				}
				end++
			}
			if end >= len(line) {
				return nil, false // unterminated: not our line to read
			}

			// A colon after the closing quote makes this a key rather than a
			// value, since a value is never followed by one.
			next := end + 1
			for next < len(line) && (line[next] == ' ' || line[next] == '\t') {
				next++
			}
			if depth != 1 || next >= len(line) || line[next] != ':' || !bytes.Equal(line[i+1:end], key) {
				i = end + 1
				continue
			}
			if seen {
				return nil, false // duplicate keys: let the decoder settle it
			}

			// Only a plain string value can be read this way.
			value := next + 1
			for value < len(line) && (line[value] == ' ' || line[value] == '\t') {
				value++
			}
			if value >= len(line) || line[value] != '"' {
				return nil, false
			}
			closing := value + 1
			for closing < len(line) && line[closing] != '"' {
				if line[closing] == '\\' {
					closing++
				}
				closing++
			}
			if closing >= len(line) {
				return nil, false
			}
			found, seen = line[value+1:closing], true
			i = closing + 1
		default:
			i++
		}
	}

	return found, seen
}

// searchable returns the bytes that must appear verbatim in a raw line for text
// to match the decoded message, or nil when there is no such sequence.
//
// Only text that JSON never escapes qualifies: `he said "hi"` is written with
// escapes in the line but not in the decoded message, so searching the raw
// bytes for it would miss real matches. The same goes for anything non-ASCII,
// which may arrive as \uXXXX. Those searches still work, they just pay for a
// full parse of every line.
func searchable(text string) []byte {
	if text == "" {
		return nil
	}

	for i := 0; i < len(text); i++ {
		if c := text[i]; c < 0x20 || c > 0x7e || c == '"' || c == '\\' {
			return nil
		}
	}

	return []byte(text)
}
