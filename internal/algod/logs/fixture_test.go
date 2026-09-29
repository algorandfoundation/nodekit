package logs

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const fixture = "testdata/node.log"

// The fixture holds real algod line shapes, including a telemetry envelope, an
// agreement entry carrying a round number, a warning whose message embeds CRLF,
// and a multi-line panic block.
func TestFixtureDefaultViewShowsProblemsOnly(t *testing.T) {
	result, err := tailFilter(fixture, 0, Filter{MinLevel: LevelWarn})
	require.NoError(t, err)

	var levels []Level
	for _, e := range result.Entries {
		levels = append(levels, e.Level)
		assert.NotEqual(t, LevelInfo, e.Level, "info must not reach the default view")
	}
	assert.Contains(t, levels, LevelWarn)
	assert.Contains(t, levels, LevelError)

	// The panic block's lines are not JSON, and are shown rather than dropped.
	assert.Contains(t, levels, LevelUnknown)
}

func TestFixtureAllLevels(t *testing.T) {
	all, err := tailFilter(fixture, 0, Filter{MinLevel: LevelTrace})
	require.NoError(t, err)

	warnOnly, err := tailFilter(fixture, 0, Filter{MinLevel: LevelWarn})
	require.NoError(t, err)

	assert.Greater(t, len(all.Entries), len(warnOnly.Entries))
}

// The flag a post-mortem reaches for. The fixture's node died of a nil
// dereference, which the Go runtime wrote as plain text, so a view that judged
// those lines by their missing level would answer this with nothing at all.
func TestFixturePanicDumpSurvivesTheTopLevels(t *testing.T) {
	for _, level := range []Level{LevelFatal, LevelPanic} {
		result, err := tailFilter(fixture, 0, Filter{MinLevel: level})
		require.NoError(t, err)

		var dump []string
		for _, e := range result.Entries {
			dump = append(dump, e.Message)
		}
		assert.Contains(t, dump, "panic: runtime error: invalid memory address or nil pointer dereference",
			"--level %s must still show the crash output", level)
	}
}

func TestFixtureEntriesAreChronological(t *testing.T) {
	result, err := tailFilter(fixture, 0, Filter{MinLevel: LevelTrace})
	require.NoError(t, err)
	require.NotEmpty(t, result.Entries)

	var previous time.Time
	for _, e := range result.Entries {
		if e.Time.IsZero() {
			continue // the panic block carries no timestamp
		}
		if !previous.IsZero() {
			assert.False(t, e.Time.Before(previous), "entries must come back oldest first")
		}
		previous = e.Time
	}
}

func TestFixtureRendersOneLinePerEntry(t *testing.T) {
	result, err := tailFilter(fixture, 0, Filter{MinLevel: LevelTrace})
	require.NoError(t, err)

	for _, e := range result.Entries {
		rendered := ansi.Strip(Render(e))
		assert.NotContains(t, rendered, "\n", "message: %q", e.Message)
	}
}

func TestFixtureKeepsRoundDropsSourceLocation(t *testing.T) {
	result, err := tailFilter(fixture, 0, Filter{MinLevel: LevelError})
	require.NoError(t, err)
	require.NotEmpty(t, result.Entries)

	var errorLine string
	for _, e := range result.Entries {
		if e.Level == LevelError {
			errorLine = ansi.Strip(Render(e))
		}
	}
	require.NotEmpty(t, errorLine)

	assert.Contains(t, errorLine, "round=48291044")
	assert.NotContains(t, errorLine, "file=")
	assert.NotContains(t, errorLine, "line=")
}

func TestFixtureTextSearch(t *testing.T) {
	result, err := tailFilter(fixture, 0, Filter{
		MinLevel: LevelTrace,
		Text:     "502",
	})
	require.NoError(t, err)
	require.Len(t, result.Entries, 1)
	assert.Contains(t, result.Entries[0].Message, "502 Bad Gateway")
}

func TestFixtureJSONOutputIsVerbatim(t *testing.T) {
	result, err := tailFilter(fixture, 0, Filter{MinLevel: LevelTrace})
	require.NoError(t, err)

	raw, err := readFixtureLines()
	require.NoError(t, err)

	for _, e := range result.Entries {
		if !e.Parsed {
			continue
		}
		assert.Contains(t, raw, string(JSONLine(e)), "--json must reproduce the original line")
	}
}

func readFixtureLines() ([]string, error) {
	data, err := os.ReadFile(fixture)
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimRight(string(data), "\n"), "\n"), nil
}
