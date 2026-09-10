package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ptr[T any](v T) *T { return &v }

// LogPaths mirrors go-algorand's (*config.Local).ResolveLogPaths. These cases
// follow its own TestResolveLogPaths so the two stay in step: the log is not
// always inside the data directory, and assuming it is would silently read the
// wrong file, or none at all.
func TestLogPaths(t *testing.T) {
	root := filepath.Join("var", "lib", "algorand")

	tests := []struct {
		name        string
		config      *Config
		wantLive    string
		wantArchive string
	}{
		{
			name:        "no config falls back to algod defaults",
			config:      nil,
			wantLive:    filepath.Join(root, "node.log"),
			wantArchive: filepath.Join(root, "node.archive.log"),
		},
		{
			name:        "empty config uses the data directory",
			config:      &Config{},
			wantLive:    filepath.Join(root, "node.log"),
			wantArchive: filepath.Join(root, "node.archive.log"),
		},
		{
			name:        "hot and cold data directories move both files",
			config:      &Config{HotDataDir: ptr("hot"), ColdDataDir: ptr("cold")},
			wantLive:    filepath.Join("hot", "node.log"),
			wantArchive: filepath.Join("cold", "node.archive.log"),
		},
		{
			name:        "LogFileDir wins over HotDataDir",
			config:      &Config{HotDataDir: ptr("hot"), LogFileDir: ptr("logs")},
			wantLive:    filepath.Join("logs", "node.log"),
			wantArchive: filepath.Join(root, "node.archive.log"),
		},
		{
			name:        "LogArchiveDir wins over ColdDataDir",
			config:      &Config{ColdDataDir: ptr("cold"), LogArchiveDir: ptr("archives")},
			wantLive:    filepath.Join(root, "node.log"),
			wantArchive: filepath.Join("archives", "node.archive.log"),
		},
		{
			name:        "custom archive name",
			config:      &Config{LogArchiveName: ptr("old.log")},
			wantLive:    filepath.Join(root, "node.log"),
			wantArchive: filepath.Join(root, "old.log"),
		},
		{
			name:        "empty strings are ignored",
			config:      &Config{HotDataDir: ptr(""), LogFileDir: ptr(""), LogArchiveName: ptr("")},
			wantLive:    filepath.Join(root, "node.log"),
			wantArchive: filepath.Join(root, "node.archive.log"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			live, archive := tc.config.LogPaths(root)
			assert.Equal(t, tc.wantLive, live)
			assert.Equal(t, tc.wantArchive, archive)
		})
	}
}

// The log is not a genesis-scoped file, unlike everything handled by
// EnsureAndResolveGenesisDirs, so no network segment is inserted.
func TestLogPathsHasNoGenesisSegment(t *testing.T) {
	live, _ := (&Config{HotDataDir: ptr("hot")}).LogPaths("root")
	assert.Equal(t, filepath.Join("hot", "node.log"), live)
	assert.NotContains(t, live, "mainnet")
}

// algod's -o flag works by forcing LogSizeLimit to zero, after which it writes
// to stdout and never creates a log file at all.
func TestLogsToStdout(t *testing.T) {
	assert.True(t, (&Config{LogSizeLimit: ptr(uint64(0))}).LogsToStdout())
	assert.False(t, (&Config{LogSizeLimit: ptr(uint64(1 << 30))}).LogsToStdout())
	assert.False(t, (&Config{}).LogsToStdout())

	var nilConfig *Config
	assert.False(t, nilConfig.LogsToStdout())
}
