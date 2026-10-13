package logs

import (
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func entryAt(level Level, msg string, ts time.Time) Entry {
	return Entry{Level: level, Message: msg, Time: ts, Parsed: true}
}

func TestFilterKeepLevels(t *testing.T) {
	levels := []Level{LevelTrace, LevelDebug, LevelInfo, LevelWarn, LevelError, LevelFatal, LevelPanic}

	// The default view: warnings and above.
	f := Filter{MinLevel: LevelWarn}
	for _, l := range levels {
		want := l >= LevelWarn
		assert.Equal(t, want, f.Keep(entryAt(l, "m", time.Time{})), "level %s", l)
	}

	// --all keeps everything.
	all := Filter{MinLevel: LevelTrace}
	for _, l := range levels {
		assert.True(t, all.Keep(entryAt(l, "m", time.Time{})), "level %s", l)
	}
}

// Unparseable lines have no level to compare, and are usually panic output, so
// no floor drops them. --level panic above all: a runtime panic is written by
// the Go runtime rather than through logrus, so the dump someone reaches that
// flag for carries no level field of its own.
func TestFilterKeepsUnknownLevelAtEveryFloor(t *testing.T) {
	unknown := Entry{Level: LevelUnknown, Message: "panic: runtime error: invalid memory address"}

	for _, level := range []Level{LevelTrace, LevelDebug, LevelInfo, LevelWarn, LevelError, LevelFatal, LevelPanic} {
		assert.True(t, Filter{MinLevel: level}.Keep(unknown), "min %s", level)
	}

	// The other filters still apply to it: only the level is exempt.
	assert.False(t, Filter{MinLevel: LevelPanic, Text: "connection reset"}.Keep(unknown))
}

func TestFilterSince(t *testing.T) {
	cutoff := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)
	f := Filter{MinLevel: LevelTrace, Since: cutoff}

	assert.False(t, f.Keep(entryAt(LevelInfo, "old", cutoff.Add(-time.Minute))))
	assert.True(t, f.Keep(entryAt(LevelInfo, "new", cutoff.Add(time.Minute))))
	assert.True(t, f.Keep(entryAt(LevelInfo, "exact", cutoff)))

	// An entry algod wrote without a usable timestamp is never dropped for
	// being too old, since we cannot know that it is.
	assert.True(t, f.Keep(entryAt(LevelInfo, "no timestamp", time.Time{})))
}

func TestFilterTextMatchesMessage(t *testing.T) {
	f := Filter{MinLevel: LevelTrace, Text: "catchup"}

	assert.True(t, f.Keep(entryAt(LevelInfo, "catchup service started", time.Time{})))
	assert.False(t, f.Keep(entryAt(LevelInfo, "unrelated", time.Time{})))

	// Matching the raw line instead of what is shown would make a search for a
	// level name match every single entry.
	self := Filter{MinLevel: LevelTrace, Text: "info"}
	e := ParseLine([]byte(`{"level":"info","msg":"nothing to see"}`))
	assert.False(t, self.Keep(e))
}

// Round numbers and addresses live in fields, and they are what people search
// for. A field is matched as decoded, unquoted key=value text.
func TestFilterTextMatchesShownFields(t *testing.T) {
	e := ParseLine([]byte(`{"Round":49291042,"file":"catchup/service.go","function":"catchup.(*Service).fetch",` +
		`"level":"info","msg":"fetched block","peer":"1.2.3.4:4160","quoted":"a \"b\" c","ok":true}`))
	shown := ansi.Strip(Render(e))

	for _, text := range []string{"49291042", "Round", "Round=49291042", "und=4929", "1.2.3.4:4160", "peer=1.2", `a "b" c`, "ok=true"} {
		assert.True(t, Filter{MinLevel: LevelTrace, Text: text}.Keep(e), "text %q, shown %q", text, shown)
	}

	// Hidden fields are not searched: the line would be kept for a reason it
	// does not show.
	for _, text := range []string{"catchup", "service.go", "file="} {
		assert.False(t, Filter{MinLevel: LevelTrace, Text: text}.Keep(e), "text %q, shown %q", text, shown)
		assert.NotContains(t, shown, text)
	}

	// Pairs are matched one at a time, so a search cannot run from one field
	// into the next.
	assert.False(t, Filter{MinLevel: LevelTrace, Text: "4160 quoted"}.Keep(e))

	// A value Render quotes is matched as decoded, not as it is quoted.
	spaced := ParseLine([]byte(`{"level":"info","msg":"m","peer":"a b"}`))
	require.Contains(t, ansi.Strip(Render(spaced)), `peer="a b"`)
	assert.True(t, Filter{MinLevel: LevelTrace, Text: "peer=a b"}.Keep(spaced))
	assert.False(t, Filter{MinLevel: LevelTrace, Text: `peer="a b"`}.Keep(spaced))
}

// Only the fields Render prints are searched, so a scalar nested in an object
// and an empty string, neither of which is shown, match nothing.
func TestFilterTextSkipsFieldsNotShown(t *testing.T) {
	e := ParseLine([]byte(`{"details":{"Round":1},"level":"info","msg":"m","x":""}`))
	shown := ansi.Strip(Render(e))

	for _, text := range []string{"Round", "details", "x=", "x"} {
		assert.False(t, Filter{MinLevel: LevelTrace, Text: text}.Keep(e), "text %q, shown %q", text, shown)
	}
}

// The text is taken literally, which is the whole point of it not being a
// regular expression: a log line is full of characters a pattern would claim.
func TestFilterTextIsLiteral(t *testing.T) {
	f := Filter{MinLevel: LevelTrace, Text: "round=4829."}
	assert.True(t, f.Keep(entryAt(LevelInfo, "sync round=4829. done", time.Time{})))
	assert.False(t, f.Keep(entryAt(LevelInfo, "sync round=48291 done", time.Time{})))

	brackets := Filter{MinLevel: LevelTrace, Text: "[::1]:4160"}
	assert.True(t, brackets.Keep(entryAt(LevelInfo, "peer [::1]:4160 left", time.Time{})))

	// Case is respected: the search is for what was typed.
	upper := Filter{MinLevel: LevelTrace, Text: "Catchup"}
	assert.False(t, upper.Keep(entryAt(LevelInfo, "catchup service started", time.Time{})))
}

func TestParseSinceDurations(t *testing.T) {
	now := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)

	got, err := ParseSince("15m", now)
	require.NoError(t, err)
	assert.Equal(t, now.Add(-15*time.Minute), got)

	got, err = ParseSince("2h", now)
	require.NoError(t, err)
	assert.Equal(t, now.Add(-2*time.Hour), got)

	// A negative duration means the same thing; nobody asks for the future.
	got, err = ParseSince("-15m", now)
	require.NoError(t, err)
	assert.Equal(t, now.Add(-15*time.Minute), got)
}

func TestParseSinceAbsolute(t *testing.T) {
	now := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		in   string
		want time.Time
	}{
		{"2026-09-09T10:30:00Z", time.Date(2026, 9, 9, 10, 30, 0, 0, time.UTC)},
		{"2026-09-08T10:30:00", time.Date(2026, 9, 8, 10, 30, 0, 0, now.Location())},
		{"2026-09-08 10:30:00", time.Date(2026, 9, 8, 10, 30, 0, 0, now.Location())},
		{"2026-09-08 10:30", time.Date(2026, 9, 8, 10, 30, 0, 0, now.Location())},
		{"2026-09-08", time.Date(2026, 9, 8, 0, 0, 0, 0, now.Location())},
		{"10:30:00", time.Date(2026, 9, 9, 10, 30, 0, 0, now.Location())},
		{"10:30", time.Date(2026, 9, 9, 10, 30, 0, 0, now.Location())},
	}

	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			got, err := ParseSince(tc.in, now)
			require.NoError(t, err)
			assert.True(t, got.Equal(tc.want), "got %s want %s", got, tc.want)
		})
	}
}

func TestParseSinceRejectsGarbage(t *testing.T) {
	now := time.Now()
	for _, in := range []string{"yesterday", "15", "next tuesday", "2026-13-45"} {
		_, err := ParseSince(in, now)
		assert.Error(t, err, "input %q", in)
	}

	empty, err := ParseSince("", now)
	require.NoError(t, err)
	assert.True(t, empty.IsZero())
}
