package logs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/algorandfoundation/nodekit/internal/algod/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func dataDirWithConfig(t *testing.T, contents string) string {
	t.Helper()
	dir := t.TempDir()
	if contents != "" {
		require.NoError(t, os.WriteFile(filepath.Join(dir, "config.json"), []byte(contents), 0o600))
	}
	return dir
}

func TestResolveSourceDefaults(t *testing.T) {
	dir := dataDirWithConfig(t, "")
	src := ResolveSource(dir)

	assert.Equal(t, dir, src.DataDir)
	assert.Equal(t, filepath.Join(dir, "node.log"), src.Path)
	assert.Equal(t, filepath.Join(dir, "node.archive.log"), src.Archive)
}

func TestResolveSourceHonoursLogFileDir(t *testing.T) {
	dir := dataDirWithConfig(t, `{"LogFileDir":"/var/log/algorand","HotDataDir":"/mnt/hot"}`)
	src := ResolveSource(dir)
	assert.Equal(t, filepath.Join("/var/log/algorand", "node.log"), src.Path)
}

// An unreadable or malformed config.json is normal for an unprivileged user and
// must never stop the command; the algod defaults are the right fallback.
func TestResolveSourceToleratesBadConfig(t *testing.T) {
	dir := dataDirWithConfig(t, `{this is not json`)
	src := ResolveSource(dir)

	assert.Nil(t, src.Config)
	assert.Equal(t, filepath.Join(dir, "node.log"), src.Path)
}

func TestSourceFloor(t *testing.T) {
	level := uint32(2)
	src := Source{Config: &config.Config{BaseLoggerDebugLevel: &level}}

	floor, ok := src.Floor()
	require.True(t, ok)
	assert.Equal(t, LevelError, floor)

	value, ok := src.FloorValue()
	require.True(t, ok)
	assert.Equal(t, uint32(2), value)

	// Without a config there is nothing to report, and guessing would produce a
	// confident but wrong message.
	_, ok = Source{}.Floor()
	assert.False(t, ok)
}

// A node below warn level never records warnings, so the command's default view
// is incomplete in a way the output alone cannot reveal.
func TestSourceHidesWarnings(t *testing.T) {
	tests := []struct {
		level uint32
		want  bool
	}{
		{0, true},  // panic only
		{1, true},  // fatal
		{2, true},  // error
		{3, false}, // warn
		{4, false}, // info, the default
		{5, false}, // debug
	}

	for _, tc := range tests {
		level := tc.level
		src := Source{Config: &config.Config{BaseLoggerDebugLevel: &level}}
		assert.Equal(t, tc.want, src.HidesWarnings(), "BaseLoggerDebugLevel %d", tc.level)
	}

	assert.False(t, Source{}.HidesWarnings(), "an unknown floor must not warn")
}

// Asking for a level the node never writes explains an empty result.
func TestSourceHides(t *testing.T) {
	level := uint32(4) // Info, algod's default
	src := Source{Config: &config.Config{BaseLoggerDebugLevel: &level}}

	assert.True(t, src.Hides(LevelDebug))
	assert.True(t, src.Hides(LevelTrace))
	assert.False(t, src.Hides(LevelInfo))
	assert.False(t, src.Hides(LevelWarn))

	assert.False(t, Source{}.Hides(LevelDebug), "an unknown floor explains nothing")
}

func TestSourceLogsToStdout(t *testing.T) {
	dir := dataDirWithConfig(t, `{"LogSizeLimit":0}`)
	assert.True(t, ResolveSource(dir).LogsToStdout())

	other := dataDirWithConfig(t, `{"LogSizeLimit":1073741824}`)
	assert.False(t, ResolveSource(other).LogsToStdout())
}
