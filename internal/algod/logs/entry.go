package logs

import (
	"bytes"
	"encoding/json"
	"time"
)

// maxLineBytes caps how much of a single log line is retained. algod normally
// writes modest lines, but a panic dump can be enormous. Overlong lines are
// truncated rather than dropped, since a giant line is usually exactly what the
// user came to read.
const maxLineBytes = 1 << 20 // 1 MiB

// truncationSuffix is appended to the message of a line that exceeded maxLineBytes.
const truncationSuffix = " …[truncated]"

// Entry is a single line of algod's node.log.
type Entry struct {
	// Time is the entry's timestamp, or the zero time when absent or unparseable.
	Time time.Time

	// Level is the parsed severity, or LevelUnknown for a line we could not read.
	Level Level

	// Message is the "msg" field, or the whole raw line when Parsed is false.
	Message string

	// Fields holds every key except time, level and msg. Numbers are kept as
	// json.Number so that large integers such as round numbers do not degrade
	// into float64 and render in scientific notation.
	Fields map[string]any

	// Raw is a copy of the original line, used verbatim by --json.
	Raw []byte

	// Parsed reports whether the line was valid JSON. Plain lines (panic stack
	// traces, early startup output) have Parsed false.
	Parsed bool
}

// ParseLine converts one line of node.log into an Entry. It never fails: a line
// that is not valid JSON becomes an Entry with Parsed false and the whole line
// as its Message, so that crash output is never silently discarded.
//
// The line need not be newline-terminated; any trailing CR/LF is trimmed.
func ParseLine(line []byte) Entry {
	line = bytes.TrimRight(line, "\r\n")

	truncated := false
	if len(line) > maxLineBytes {
		line = line[:maxLineBytes]
		truncated = true
	}

	// Copy the line: callers reuse their read buffers.
	raw := make([]byte, len(line))
	copy(raw, line)

	e := Entry{Level: LevelUnknown, Raw: raw}

	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		e.Message = string(trimmed)
		if truncated {
			e.Message += truncationSuffix
		}
		return e
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(trimmed, &fields); err != nil {
		e.Message = string(trimmed)
		if truncated {
			e.Message += truncationSuffix
		}
		return e
	}
	e.Parsed = true

	if v, ok := fields["time"]; ok {
		var s string
		if json.Unmarshal(v, &s) == nil {
			// A malformed timestamp leaves Time zero rather than failing the line.
			e.Time, _ = time.Parse(time.RFC3339Nano, s)
		}
		delete(fields, "time")
	}

	if v, ok := fields["level"]; ok {
		var s string
		if json.Unmarshal(v, &s) == nil {
			// An unrecognised level leaves LevelUnknown, which is never hidden
			// by severity filtering alone.
			if lvl, err := ParseLevel(s); err == nil {
				e.Level = lvl
			}
		}
		delete(fields, "level")
	}

	if v, ok := fields["msg"]; ok {
		_ = json.Unmarshal(v, &e.Message)
		delete(fields, "msg")
	}

	if len(fields) > 0 {
		e.Fields = make(map[string]any, len(fields))
		for k, v := range fields {
			dec := json.NewDecoder(bytes.NewReader(v))
			// UseNumber keeps "line":257 and "Round":48291043 as exact numbers.
			// Without it they decode to float64 and render as 4.8291043e+07.
			dec.UseNumber()
			var val any
			if dec.Decode(&val) == nil {
				e.Fields[k] = val
			}
		}
	}

	if truncated {
		e.Message += truncationSuffix
	}
	return e
}
