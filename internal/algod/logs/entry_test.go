package logs

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLine(t *testing.T) {
	tests := []struct {
		name       string
		line       string
		wantParsed bool
		wantLevel  Level
		wantMsg    string
	}{
		{
			name:       "algod warning",
			line:       `{"file":"service.go","level":"warning","line":852,"msg":"could not acquire block","time":"2026-09-09T14:16:19.773408+02:00"}`,
			wantParsed: true,
			wantLevel:  LevelWarn,
			wantMsg:    "could not acquire block",
		},
		{
			name:       "missing msg",
			line:       `{"level":"info","time":"2026-09-09T14:16:19.773408+02:00"}`,
			wantParsed: true,
			wantLevel:  LevelInfo,
			wantMsg:    "",
		},
		{
			name:       "unrecognised level is unknown, not an error",
			line:       `{"level":"verbose","msg":"hello"}`,
			wantParsed: true,
			wantLevel:  LevelUnknown,
			wantMsg:    "hello",
		},
		{
			name:       "missing level",
			line:       `{"msg":"hello"}`,
			wantParsed: true,
			wantLevel:  LevelUnknown,
			wantMsg:    "hello",
		},
		{
			name:       "plain text line is kept whole",
			line:       `panic: runtime error: index out of range`,
			wantParsed: false,
			wantLevel:  LevelUnknown,
			wantMsg:    "panic: runtime error: index out of range",
		},
		{
			name:       "malformed json falls back to plain",
			line:       `{"level":"info", "msg":`,
			wantParsed: false,
			wantLevel:  LevelUnknown,
			wantMsg:    `{"level":"info", "msg":`,
		},
		{
			name:       "json array is not an entry",
			line:       `["not","an","entry"]`,
			wantParsed: false,
			wantLevel:  LevelUnknown,
			wantMsg:    `["not","an","entry"]`,
		},
		{
			name:       "escaped newlines survive parsing",
			line:       `{"level":"warning","msg":"502 Bad Gateway\r\nopenresty"}`,
			wantParsed: true,
			wantLevel:  LevelWarn,
			wantMsg:    "502 Bad Gateway\r\nopenresty",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := ParseLine([]byte(tc.line))
			assert.Equal(t, tc.wantParsed, e.Parsed)
			assert.Equal(t, tc.wantLevel, e.Level)
			assert.Equal(t, tc.wantMsg, e.Message)
		})
	}
}

func TestParseLineTrimsLineEndings(t *testing.T) {
	e := ParseLine([]byte(`{"level":"info","msg":"hi"}` + "\r\n"))
	require.True(t, e.Parsed)
	assert.Equal(t, "hi", e.Message)
	assert.Equal(t, `{"level":"info","msg":"hi"}`, string(e.Raw))
}

func TestParseLineTime(t *testing.T) {
	e := ParseLine([]byte(`{"level":"info","msg":"hi","time":"2026-09-09T14:16:19.773408+02:00"}`))
	require.True(t, e.Parsed)
	assert.Equal(t, 2026, e.Time.Year())
	assert.Equal(t, time.September, e.Time.Month())

	// A malformed timestamp must not discard the entry.
	bad := ParseLine([]byte(`{"level":"info","msg":"hi","time":"not-a-time"}`))
	require.True(t, bad.Parsed)
	assert.True(t, bad.Time.IsZero())
	assert.Equal(t, "hi", bad.Message)
}

// Large integers must stay exact. Decoded as float64 they would render as
// 4.8291043e+07, which is visible in the default output because round numbers
// are among the fields that are kept.
func TestParseLineKeepsNumbersExact(t *testing.T) {
	e := ParseLine([]byte(`{"level":"info","msg":"hi","Round":48291043,"line":257}`))
	require.True(t, e.Parsed)

	round, ok := e.Fields["Round"].(json.Number)
	require.True(t, ok, "Round should decode as json.Number, got %T", e.Fields["Round"])
	assert.Equal(t, "48291043", round.String())

	line, ok := e.Fields["line"].(json.Number)
	require.True(t, ok)
	assert.Equal(t, "257", line.String())
}

func TestParseLineSeparatesKnownFields(t *testing.T) {
	e := ParseLine([]byte(`{"level":"warning","msg":"m","time":"2026-09-09T14:16:19+02:00","Context":"sync"}`))
	require.True(t, e.Parsed)
	assert.NotContains(t, e.Fields, "level")
	assert.NotContains(t, e.Fields, "msg")
	assert.NotContains(t, e.Fields, "time")
	assert.Contains(t, e.Fields, "Context")
}

func TestParseLineTruncatesOverlongLines(t *testing.T) {
	huge := `{"level":"info","msg":"` + strings.Repeat("x", 2<<20) + `"}`
	e := ParseLine([]byte(huge))

	// Truncation breaks the JSON, so the line degrades to a plain entry, but it
	// is kept rather than dropped: a giant line is usually a panic dump.
	assert.False(t, e.Parsed)
	assert.True(t, strings.HasSuffix(e.Message, truncationSuffix))
	assert.LessOrEqual(t, len(e.Raw), maxLineBytes)
}

// Raw must be a copy: the scan and Follow both reuse their read buffers.
func TestParseLineCopiesRaw(t *testing.T) {
	buf := []byte(`{"level":"info","msg":"first"}`)
	e := ParseLine(buf)
	copy(buf, []byte(`{"level":"info","msg":"XXXXX"}`))
	assert.Equal(t, `{"level":"info","msg":"first"}`, string(e.Raw))
}
