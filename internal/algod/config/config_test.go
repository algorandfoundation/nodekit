package config

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

// IsEqual and MergeAlgodConfigs have to know about every field of Config. One
// they do not handle makes two configs that differ compare equal, so `configure
// algod` decides there is nothing to do, and a value passed to the merge is
// dropped before it can be written.
//
// The fields are walked by reflection rather than listed, so that a field added
// to Config is covered here without anyone having to remember this test.
func TestIsEqualAndMergeCoverEveryField(t *testing.T) {
	fields := reflect.VisibleFields(reflect.TypeOf(Config{}))
	require.NotEmpty(t, fields)

	for _, field := range fields {
		t.Run(field.Name, func(t *testing.T) {
			require.Equal(t, reflect.Pointer, field.Type.Kind(),
				"every field is optional, so that absent and set to the zero value stay distinct")

			// Set to a pointer to the zero value: absent and present-but-zero
			// are different configs, and algod's defaults are not all zero.
			var set Config
			reflect.ValueOf(&set).Elem().FieldByIndex(field.Index).Set(reflect.New(field.Type.Elem()))

			var absent Config
			assert.False(t, absent.IsEqual(set), "IsEqual does not compare %s", field.Name)
			assert.False(t, set.IsEqual(absent), "IsEqual does not compare %s", field.Name)

			merged := MergeAlgodConfigs(absent, set)
			assert.False(t, reflect.ValueOf(merged).FieldByIndex(field.Index).IsNil(),
				"MergeAlgodConfigs drops %s", field.Name)
			assert.True(t, merged.IsEqual(set))

			kept := MergeAlgodConfigs(set, absent)
			assert.True(t, kept.IsEqual(set), "a field the override leaves unset must keep its value")
		})
	}
}

// The shape `configure algod` uses: the current config on disk, and an override
// built from the flags the user actually passed.
func TestMergeAlgodConfigsOverridesOnlyWhatIsSet(t *testing.T) {
	current := Config{EnableP2P: ptr(true), LogFileDir: ptr(filepath.Join("var", "log", "algorand"))}
	override := Config{LogFileDir: ptr(filepath.Join("mnt", "logs"))}

	merged := MergeAlgodConfigs(current, override)
	assert.Equal(t, filepath.Join("mnt", "logs"), *merged.LogFileDir)
	require.NotNil(t, merged.EnableP2P)
	assert.True(t, *merged.EnableP2P, "a field the override leaves unset keeps its value")
	assert.False(t, current.IsEqual(merged), "a config that differs only in a log path is not up to date")
}
