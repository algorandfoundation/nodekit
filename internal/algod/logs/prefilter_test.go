package logs

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPeekLevelReadsTheWireSpellings(t *testing.T) {
	cases := map[string]Level{
		`{"level":"trace","msg":"m"}`:   LevelTrace,
		`{"level":"debug","msg":"m"}`:   LevelDebug,
		`{"level":"info","msg":"m"}`:    LevelInfo,
		`{"level":"warning","msg":"m"}`: LevelWarn,
		`{"level":"warn","msg":"m"}`:    LevelWarn,
		`{"level":"error","msg":"m"}`:   LevelError,
		`{"level":"fatal","msg":"m"}`:   LevelFatal,
		`{"level":"panic","msg":"m"}`:   LevelPanic,

		// The level is rarely the first key: sorted output puts capitalised
		// keys, "file" and "function" ahead of it.
		`{"Context":"sync","file":"s.go","level":"error","msg":"m"}`: LevelError,

		// A telemetry envelope nests an object under "details", which sorts
		// first. The entry's own level is the one at the outer level.
		`{"details":{"level":"info"},"level":"error","msg":"m"}`: LevelError,

		// The key spelled out inside a message is not a key.
		`{"level":"error","msg":"saw \"level\":\"info\" in a payload"}`: LevelError,
	}

	for line, want := range cases {
		level, ok := peekLevel([]byte(line))
		require.True(t, ok, "line %s", line)
		assert.Equal(t, want, level, "line %s", line)
	}
}

// Every one of these must fall through to a full parse. Reporting a level for
// any of them would let the prescreen discard a line Filter.Keep would show.
func TestPeekLevelDeclinesWhatItCannotBeSureOf(t *testing.T) {
	lines := []string{
		// Not JSON at all. A panic dump can quote anything, including this key.
		`goroutine 1 [running]: boom ,"level":"info", boom`,
		`panic: runtime error`,
		``,

		// No level field.
		`{"msg":"m","time":"2026-09-09T14:16:19+02:00"}`,
		`{}`,

		// A level algod does not write.
		`{"level":"notice","msg":"m"}`,

		// Only a nested level, so the entry has none of its own. Both of these
		// came out of the fuzzer, which is where the seed corpus keeps them.
		`{"":{"level":"info"}}`,
		`{"details":{"level":"info"},"msg":"m"}`,
		`{"Overrides":[{"level":"info"}],"msg":"m"}`,

		// Torn after the level field: enough to read a level from, not enough
		// for ParseLine to decode, so ParseLine shows it as an unknown level.
		`{"level":"info"connection reset`,
		`{"level":"info"}}`,

		// Truncated mid-line by the 1 MiB cap, or caught mid-write.
		`{"file":"s.go","level":"in`,
		`{"level":`,

		// Duplicate keys: which one wins is the decoder's business.
		`{"level":"info","level":"error","msg":"m"}`,

		// Not a string.
		`{"level":3,"msg":"m"}`,
	}

	for _, line := range lines {
		_, ok := peekLevel([]byte(line))
		assert.False(t, ok, "must not decide: %s", line)
	}

	// Past the cap ParseLine truncates the line and the remains no longer
	// decode, so the level visible here is not the level the entry will have.
	huge := `{"level":"info","msg":"` + strings.Repeat("x", maxLineBytes) + `"}`
	_, ok := peekLevel([]byte(huge))
	assert.False(t, ok, "must not decide a line ParseLine will truncate")
}

func TestSearchableOnlyAcceptsEscapeFreeText(t *testing.T) {
	// Pattern syntax is not special here: these are ordinary bytes to look for.
	usable := []string{"connection reset", "502", "round=48291043", "^peer", `warn|error`, "(?i)round"}
	for _, text := range usable {
		assert.Equal(t, []byte(text), searchable(text), "text %q", text)
	}

	// Text that JSON may escape would be searched for in a form the raw line
	// does not contain, so real matches would be dropped. Those searches still
	// work; they just fall through to the full parse.
	unusable := []string{
		`he said "hi"`, // quotes are escaped in the line
		`C:\path`,      // so are backslashes, including one in a would-be pattern
		`round=\d+`,
		"genesisID=Ω", // non-ASCII may arrive as \uXXXX
		"a\tb",        // so do control characters
		"<nil>",       // HTML escaping writes < and > as \u003c and \u003e
		"a & b",       // and & as \u0026
		``,
	}
	for _, text := range unusable {
		assert.Nil(t, searchable(text), "text %q", text)
	}
}

// prescreenCorpus is deliberately adversarial: the shapes that could trick a
// byte-level scan sit next to the ordinary ones.
var prescreenCorpus = []string{
	`{"file":"s.go","level":"info","line":324,"msg":"Sync round set to 48291043","time":"2026-09-09T14:16:10.1+02:00"}`,
	`{"file":"s.go","level":"warning","msg":"fetchRound could not acquire block","time":"2026-09-09T14:16:11+02:00"}`,
	`{"Context":"sync","level":"error","msg":"peer 1.2.3.4:4160 error: connection reset by peer","round":48291044,"time":"2026-09-09T14:16:12+02:00"}`,
	`{"level":"fatal","msg":"cannot open ledger","time":"2026-09-09T14:16:13+02:00"}`,
	`{"level":"panic","msg":"assertion failed","time":"2026-09-09T14:16:14+02:00"}`,
	`{"level":"debug","msg":"noise"}`,
	`{"level":"trace","msg":"noise"}`,
	`{"level":"warn","msg":"short spelling"}`,
	`{"level":"notice","msg":"a level algod does not write"}`,
	`{"msg":"no level at all","time":"2026-09-09T14:16:15+02:00"}`,
	`{"details":{"level":"info","Version":"5.0.0"},"level":"error","msg":"telemetry envelope"}`,
	`{"level":"info","msg":"quoting \"level\":\"error\" inside the message"}`,
	`{"level":"info","msg":"he said \"connection reset\" once"}`,
	`{"level":"error","msg":"unicode \u03a9 and a tab \t"}`,
	// logrus leaves Go's HTML escaping on, so these three bytes never
	// appear verbatim in a line however plain the message reads.
	`{"level":"error","msg":"peer \u003cnil\u003e reset \u0026 dropped"}`,
	`goroutine 1 [running]:`,
	`panic: runtime error: index out of range ,"level":"info",`,
	`  {"level":"info","msg":"leading whitespace"}`,
	`{"level":"in`,
	`{"level":"info"connection reset`,
	`{"":{"level":"info"}}`,
	`{"Overrides":[{"level":"info"}],"msg":"m"}`,
	`{"level":"info","level":"error","msg":"duplicate keys"}`,
	`{"level":3,"msg":"not a string"}`,
	`{}`,
	``,
}

// The prescreen is one-sided: rejecting a line the full filter would have kept
// is the only way it can be wrong, and it loses log entries silently when it is.
func TestPrescreenNeverRejectsAKeptLine(t *testing.T) {
	filters := []Filter{
		{MinLevel: LevelWarn},
		{MinLevel: LevelTrace},
		{MinLevel: LevelError},
		{MinLevel: LevelPanic},
		{MinLevel: LevelWarn, Text: "connection reset"},
		{MinLevel: LevelTrace, Text: "connection reset"},
		{MinLevel: LevelTrace, Text: "level"},
		{MinLevel: LevelTrace, Text: "panic"},
		{MinLevel: LevelTrace, Text: `he said "connection reset"`},
		{MinLevel: LevelTrace, Text: "unicode Ω"},
		{MinLevel: LevelTrace, Text: "<nil>"},
		{MinLevel: LevelTrace, Text: "reset & dropped"},
	}

	for _, f := range filters {
		pre := newPrescreen(f)
		for _, line := range prescreenCorpus {
			if !pre.rejects([]byte(line)) {
				continue
			}
			assert.False(t, f.Keep(ParseLine([]byte(line))),
				"prescreen dropped a line the filter keeps\n filter: %+v\n line: %s", f, line)
		}
	}
}

// The other way it can be wrong is by never rejecting anything, which costs
// nothing in correctness and everything in the speed it exists for.
func TestPrescreenRejectsTheCommonCase(t *testing.T) {
	pre := newPrescreen(Filter{MinLevel: LevelWarn})
	assert.True(t, pre.rejects([]byte(`{"file":"s.go","level":"info","msg":"Sync round set to 48291043"}`)))
	assert.False(t, pre.rejects([]byte(`{"file":"s.go","level":"warning","msg":"fetchRound could not acquire block"}`)))

	text := newPrescreen(Filter{MinLevel: LevelTrace, Text: "connection reset"})
	assert.True(t, text.rejects([]byte(`{"level":"info","msg":"Sync round set to 48291043"}`)))
	assert.False(t, text.rejects([]byte(`{"level":"error","msg":"peer error: connection reset by peer"}`)))
}

// A filtered read must return exactly what a straight forward scan of the file
// would, prescreen or no prescreen.
func TestTailFilterMatchesUnfilteredScanOfFixture(t *testing.T) {
	withChunkSize(t, 128) // force the chunk-boundary carry to run too

	raw, err := readFixtureLines()
	require.NoError(t, err)

	filters := []Filter{
		{MinLevel: LevelWarn},
		{MinLevel: LevelTrace},
		{MinLevel: LevelError},
		{MinLevel: LevelTrace, Text: "502"},
		{MinLevel: LevelTrace, Text: "round"},
		{MinLevel: LevelWarn, Text: "Bad Gateway"},
	}

	for _, f := range filters {
		want := []string{} // messages() never returns nil, so match its shape
		for _, line := range raw {
			if strings.TrimSpace(line) == "" {
				continue
			}
			if e := ParseLine([]byte(line)); f.Keep(e) {
				want = append(want, e.Message)
			}
		}

		// n of 0 streams forwards and a count walks backwards, and the
		// prescreen sits in both.
		for _, n := range []int{0, 1000} {
			result, err := tailFilter(fixture, n, f)
			require.NoError(t, err)
			assert.Equal(t, want, messages(result.Entries), "filter %+v, n %d", f, n)
		}
	}
}

func FuzzPrescreenNeverRejectsAKeptLine(f *testing.F) {
	for _, line := range prescreenCorpus {
		f.Add(line, uint8(LevelWarn), "connection reset")
	}
	f.Add(`{"level":"info","msg":"x"}`, uint8(LevelTrace), `"`)

	f.Fuzz(func(t *testing.T, line string, minLevel uint8, text string) {
		if strings.ContainsAny(line, "\n\r") {
			t.Skip("the reader only ever hands whole lines to the prescreen")
		}
		if minLevel > uint8(LevelPanic) {
			t.Skip()
		}
		filter := Filter{MinLevel: Level(minLevel), Text: text}

		if !newPrescreen(filter).rejects([]byte(line)) {
			return
		}
		if filter.Keep(ParseLine([]byte(line))) {
			t.Fatalf("prescreen dropped a line the filter keeps: %q (filter %+v)", line, filter)
		}
	})
}
