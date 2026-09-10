package logs

import (
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// plain renders an entry with any styling removed, so assertions compare text
// rather than escape sequences. lipgloss already degrades to plain output when
// stdout is not a terminal; stripping makes the tests independent of that.
func plain(e Entry) string {
	return ansi.Strip(Render(e))
}

func at(t *testing.T, value string) time.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	require.NoError(t, err)
	return parsed
}

func TestRenderLine(t *testing.T) {
	e := ParseLine([]byte(`{"level":"warning","msg":"sync fell behind","time":"2026-09-09T14:16:19+02:00"}`))
	got := plain(e)

	assert.Contains(t, got, "WARN")
	assert.Contains(t, got, "sync fell behind")
	assert.Contains(t, got, e.Time.Local().Format(timeLayout))
}

func TestRenderMissingTimestampKeepsAlignment(t *testing.T) {
	got := plain(Entry{Level: LevelError, Message: "no time"})
	assert.Contains(t, got, noTime)
}

func TestRenderLevelIsPadded(t *testing.T) {
	// "WARN" pads to the width of "ERROR" so messages line up.
	assert.Contains(t, plain(Entry{Level: LevelWarn, Message: "m"}), "WARN  m")
	assert.Contains(t, plain(Entry{Level: LevelError, Message: "m"}), "ERROR m")
	assert.Contains(t, plain(Entry{Level: LevelUnknown, Message: "m"}), "????? m")
}

// Code locations and telemetry plumbing appear on nearly every algod entry and
// say nothing about the node, but semantic fields such as round must survive.
func TestRenderDropsNoiseFieldsAndKeepsSemanticOnes(t *testing.T) {
	e := ParseLine([]byte(`{"level":"warning","msg":"stale request",` +
		`"file":"wsPeer.go","function":"readLoop","line":565,"name":"","Context":"sync",` +
		`"instanceName":"abc","session":"s","v":"5.0.0",` +
		`"round":48291043,"peer":"10.0.0.4"}`))
	got := plain(e)

	for _, hidden := range []string{"file=", "function=", "line=", "name=", "Context=", "instanceName=", "session=", "v="} {
		assert.NotContains(t, got, hidden)
	}
	assert.Contains(t, got, "round=48291043")
	assert.Contains(t, got, "peer=10.0.0.4")
}

func TestRenderSkipsNestedValues(t *testing.T) {
	// Telemetry envelopes are objects; skipping non-scalars drops them without
	// having to name every one.
	e := ParseLine([]byte(`{"level":"info","msg":"startup",` +
		`"details":{"Version":"5.0.0"},"Overrides":[1,2],"round":7}`))
	got := plain(e)

	assert.NotContains(t, got, "details=")
	assert.NotContains(t, got, "Overrides=")
	assert.Contains(t, got, "round=7")
}

func TestRenderSkipsEmptyValues(t *testing.T) {
	e := ParseLine([]byte(`{"level":"info","msg":"m","empty":"","kept":"yes"}`))
	got := plain(e)
	assert.NotContains(t, got, "empty=")
	assert.Contains(t, got, "kept=yes")
}

func TestRenderFieldsAreSorted(t *testing.T) {
	e := ParseLine([]byte(`{"level":"info","msg":"m","zebra":1,"alpha":2,"middle":3}`))
	got := plain(e)
	assert.Less(t, strings.Index(got, "alpha="), strings.Index(got, "middle="))
	assert.Less(t, strings.Index(got, "middle="), strings.Index(got, "zebra="))
}

func TestRenderQuotesValuesNeedingIt(t *testing.T) {
	e := ParseLine([]byte(`{"level":"info","msg":"m","addr":"with space","plain":"none"}`))
	got := plain(e)
	assert.Contains(t, got, `addr="with space"`)
	assert.Contains(t, got, "plain=none")
}

// One entry must stay one line, or the output stops being greppable. algod's
// 502 warnings carry a CRLF-laden HTML body in the message.
func TestRenderCollapsesNewlines(t *testing.T) {
	e := ParseLine([]byte(`{"level":"warning","msg":"502 Bad Gateway\r\n<html>\n</html>"}`))
	got := plain(e)
	assert.NotContains(t, got, "\n")
	assert.NotContains(t, got, "\r")
	assert.Contains(t, got, "502 Bad Gateway")
}

func TestRenderLargeNumbersAreNotScientific(t *testing.T) {
	e := ParseLine([]byte(`{"level":"info","msg":"m","Round":48291043}`))
	got := plain(e)
	assert.Contains(t, got, "Round=48291043")
	assert.NotContains(t, got, "e+")
}

// --json must be byte-identical to what algod wrote, so that reading the file
// directly and piping this command to jq agree.
func TestJSONLinePassesParsedEntriesThrough(t *testing.T) {
	raw := `{"level":"warning","msg":"hi","time":"2026-09-09T14:16:19+02:00","round":7}`
	e := ParseLine([]byte(raw))
	require.True(t, e.Parsed)
	assert.Equal(t, raw, string(JSONLine(e)))
}

// Plain lines have no JSON form of their own, but the stream must stay valid
// newline-delimited JSON.
func TestJSONLineWrapsPlainLines(t *testing.T) {
	e := ParseLine([]byte(`panic: runtime error`))
	require.False(t, e.Parsed)
	assert.JSONEq(t, `{"level":"unknown","msg":"panic: runtime error"}`, string(JSONLine(e)))
}

func TestRenderUsesLocalTime(t *testing.T) {
	e := Entry{Level: LevelInfo, Message: "m", Time: at(t, "2026-09-09T14:16:19+02:00")}
	assert.Contains(t, plain(e), e.Time.Local().Format(timeLayout))
}
