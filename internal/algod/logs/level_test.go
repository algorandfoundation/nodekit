package logs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// logrus marshals WarnLevel as "warning", so a parser that only knows "warn"
// silently matches nothing and the default view of the command comes back
// empty. This is the most important case in the package.
func TestParseLevelAcceptsLogrusWarningSpelling(t *testing.T) {
	for _, spelling := range []string{"warning", "Warning", "WARNING", "warn", "WARN"} {
		level, err := ParseLevel(spelling)
		require.NoError(t, err, "spelling %q", spelling)
		assert.Equal(t, LevelWarn, level, "spelling %q", spelling)
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		in      string
		want    Level
		wantErr bool
	}{
		{in: "trace", want: LevelTrace},
		{in: "debug", want: LevelDebug},
		{in: "info", want: LevelInfo},
		{in: "INFO", want: LevelInfo},
		{in: " info ", want: LevelInfo},
		{in: "warn", want: LevelWarn},
		{in: "warning", want: LevelWarn},
		{in: "error", want: LevelError},
		{in: "err", want: LevelError},
		{in: "fatal", want: LevelFatal},
		{in: "panic", want: LevelPanic},
		{in: "", wantErr: true},
		{in: "verbose", wantErr: true},
		{in: "5", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			got, err := ParseLevel(tc.in)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestLevelOrdering(t *testing.T) {
	// Ascending severity is what makes filtering a single >= comparison.
	assert.True(t, LevelTrace < LevelDebug)
	assert.True(t, LevelDebug < LevelInfo)
	assert.True(t, LevelInfo < LevelWarn)
	assert.True(t, LevelWarn < LevelError)
	assert.True(t, LevelError < LevelFatal)
	assert.True(t, LevelFatal < LevelPanic)
}

func TestLevelString(t *testing.T) {
	assert.Equal(t, "warn", LevelWarn.String())
	assert.Equal(t, "error", LevelError.String())
	assert.Equal(t, "unknown", LevelUnknown.String())
}

// BaseLoggerDebugLevel uses go-algorand's logging.Level numbering, which runs
// Panic=0 .. Debug=5 and has no trace level at all.
func TestAlgodLevel(t *testing.T) {
	tests := []struct {
		in   AlgodLevel
		want Level
	}{
		{0, LevelPanic},
		{1, LevelFatal},
		{2, LevelError},
		{3, LevelWarn},
		{4, LevelInfo},
		{5, LevelDebug},
		{6, LevelDebug},
	}
	for _, tc := range tests {
		assert.Equal(t, tc.want, tc.in.Level(), "AlgodLevel(%d)", tc.in)
	}
}
