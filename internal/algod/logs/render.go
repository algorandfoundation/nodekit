package logs

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/algorandfoundation/nodekit/ui/style"
)

// timeLayout is the timestamp shown on each rendered line. Seconds resolution
// and no date keeps lines narrow; --json carries the full RFC3339 timestamp for
// anyone who needs it.
const timeLayout = "15:04:05"

// noTime is shown in place of a timestamp algod did not write, keeping the
// columns aligned.
const noTime = "--:--:--"

// levelWidth pads the level column so messages line up.
const levelWidth = 5

// hiddenFields are dropped from the rendered line. They are Go source
// locations and telemetry plumbing that appear on nearly every entry and say
// nothing about the node; --json still exposes them.
var hiddenFields = map[string]bool{
	"file":         true,
	"function":     true,
	"line":         true,
	"name":         true,
	"Context":      true,
	"instanceName": true,
	"session":      true,
	"v":            true,
}

// Render formats an entry as a single human-readable line:
//
//	[15:04:05] WARN  message key=value
//
// Colour is applied to the timestamp and level only. lipgloss degrades to plain
// text automatically when stdout is not a terminal, so no TTY check is needed.
func Render(e Entry) string {
	stamp := noTime
	if !e.Time.IsZero() {
		stamp = e.Time.Local().Format(timeLayout)
	}

	var b strings.Builder
	b.WriteString(style.Blue.Render("[" + stamp + "]"))
	b.WriteByte(' ')
	b.WriteString(renderLevel(e.Level))
	b.WriteByte(' ')
	b.WriteString(collapseNewlines(e.Message))

	if fields := renderFields(e.Fields); fields != "" {
		b.WriteByte(' ')
		b.WriteString(style.Blue.Render(fields))
	}

	return b.String()
}

// renderLevel returns the padded, coloured level token.
func renderLevel(l Level) string {
	name := strings.ToUpper(l.String())
	if l == LevelUnknown {
		name = "?????"
	}
	if len(name) < levelWidth {
		name += strings.Repeat(" ", levelWidth-len(name))
	}

	switch l {
	case LevelPanic, LevelFatal, LevelError:
		return style.Red.Render(name)
	case LevelWarn:
		return style.Yellow.Render(name)
	case LevelInfo:
		return style.Cyan.Render(name)
	case LevelDebug, LevelTrace:
		return style.Blue.Render(name)
	default:
		return name
	}
}

// renderFields formats the entry's remaining fields as sorted key=value pairs.
//
// Only scalars are shown: skipping objects and arrays drops algod's nested
// telemetry envelopes without needing to name them. Empty strings are skipped
// too, since algod writes "name":"" on a large share of entries.
func renderFields(fields map[string]any) string {
	if len(fields) == 0 {
		return ""
	}

	keys := make([]string, 0, len(fields))
	for k := range fields {
		if hiddenFields[k] {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		value, ok := scalar(fields[k])
		if !ok || value == "" {
			continue
		}
		parts = append(parts, k+"="+value)
	}
	return strings.Join(parts, " ")
}

// scalar renders a field value, reporting false for objects, arrays and nulls.
// Strings containing spaces or equals signs are quoted so pairs stay separable,
// and strings containing a control character are quoted so that one entry stays
// one line.
//
// The newline case is the reason the test is not only about separability. No
// field algod writes to node.log carries a newline today: a stack trace goes
// into the message, where collapseNewlines already handles it, and the one
// structurally multi-line field is built by a telemetry hook on a copy of the
// entry that never reaches the file. But Render's one-line guarantee is a
// property of Render, not of algod's vocabulary, and --file points the command
// at whatever file the user names. %q escapes CR, LF, quotes and backslashes
// together, so the guarantee stops depending on the producer.
func scalar(v any) (string, bool) {
	switch t := v.(type) {
	case string:
		if strings.ContainsAny(t, " =\"\\\r\n") {
			return fmt.Sprintf("%q", t), true
		}
		return t, true
	case json.Number:
		return t.String(), true
	case bool:
		return fmt.Sprintf("%t", t), true
	default:
		return "", false
	}
}

// collapseNewlines flattens an embedded newline so one entry stays one line and
// the output remains greppable. algod escapes CRLF inside JSON strings, so this
// applies once the message has been unescaped.
func collapseNewlines(s string) string {
	if !strings.ContainsAny(s, "\r\n") {
		return s
	}
	replacer := strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ")
	return replacer.Replace(s)
}

// JSONLine returns the entry as a single JSON line for --json output.
//
// Parsed entries are passed through byte for byte, so that piping to jq yields
// exactly what algod wrote, with no field reordering and nothing lost to a
// round trip through a Go type. Lines that were not JSON have no original form,
// so they are wrapped to keep the stream valid newline-delimited JSON.
func JSONLine(e Entry) []byte {
	if e.Parsed {
		return e.Raw
	}
	wrapped, err := json.Marshal(struct {
		Level string `json:"level"`
		Msg   string `json:"msg"`
	}{Level: LevelUnknown.String(), Msg: e.Message})
	if err != nil {
		// Marshalling a struct of two strings cannot fail, but never emit a
		// broken line if it somehow does.
		return []byte(`{"level":"unknown","msg":""}`)
	}
	return wrapped
}
