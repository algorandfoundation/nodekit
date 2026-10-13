package config

import "path/filepath"

// Config represents the configuration settings for algod, including enabling P2PHybrid
//
// This is a partial model of go-algorand's config.Local: only the fields nodekit
// needs are declared, and unknown keys in config.json are ignored.
type Config struct {
	EnableP2P           *bool `json:"EnableP2P,omitempty"`
	EnableP2PHybridMode *bool `json:"EnableP2PHybridMode,omitempty"`

	// BaseLoggerDebugLevel is the severity floor algod applies to node.log.
	// It uses go-algorand's logging.Level numbering, Panic=0 .. Debug=5, and
	// defaults to 4 (Info). Entries below it are never written at all.
	BaseLoggerDebugLevel *uint32 `json:"BaseLoggerDebugLevel,omitempty"`

	// LogSizeLimit is the size at which node.log is rotated to the archive.
	// It defaults to 1 GiB. Zero is special: algod then logs to stdout and
	// never creates node.log. The algod -o flag works by forcing this to 0.
	LogSizeLimit *uint64 `json:"LogSizeLimit,omitempty"`

	// HotDataDir, ColdDataDir, LogFileDir and LogArchiveDir relocate the log
	// files away from the data directory. See ResolveLogPaths in
	// go-algorand's config/localTemplate.go, which LogPath mirrors.
	HotDataDir    *string `json:"HotDataDir,omitempty"`
	ColdDataDir   *string `json:"ColdDataDir,omitempty"`
	LogFileDir    *string `json:"LogFileDir,omitempty"`
	LogArchiveDir *string `json:"LogArchiveDir,omitempty"`

	// LogArchiveName is the filename of the rotated log, default node.archive.log.
	LogArchiveName *string `json:"LogArchiveName,omitempty"`
}

// LiveLogName is the fixed filename algod uses for the active log.
const LiveLogName = "node.log"

// DefaultLogArchiveName is algod's default for LogArchiveName.
const DefaultLogArchiveName = "node.archive.log"

// DefaultBaseLoggerDebugLevel is algod's default BaseLoggerDebugLevel, Info.
const DefaultBaseLoggerDebugLevel = uint32(4)

// LogPaths mirrors (*config.Local).ResolveLogPaths from go-algorand
// (config/localTemplate.go), returning the locations algod would use for the
// live log and its archive given this config and a data directory.
//
// The precedence is significant and the last match wins: rootDir, then
// HotDataDir, then LogFileDir for the live log; rootDir, then ColdDataDir, then
// LogArchiveDir for the archive. Each is joined to the bare directory, with no
// genesis ID segment.
//
// A nil Config yields algod's defaults, which is the correct fallback when
// config.json is missing or unreadable.
func (c *Config) LogPaths(rootDir string) (liveLog string, archive string) {
	archiveName := DefaultLogArchiveName
	if c != nil && c.LogArchiveName != nil && *c.LogArchiveName != "" {
		archiveName = *c.LogArchiveName
	}

	liveLog = filepath.Join(rootDir, LiveLogName)
	archive = filepath.Join(rootDir, archiveName)
	if c == nil {
		return liveLog, archive
	}

	if c.HotDataDir != nil && *c.HotDataDir != "" {
		liveLog = filepath.Join(*c.HotDataDir, LiveLogName)
	}
	if c.ColdDataDir != nil && *c.ColdDataDir != "" {
		archive = filepath.Join(*c.ColdDataDir, archiveName)
	}
	if c.LogFileDir != nil && *c.LogFileDir != "" {
		liveLog = filepath.Join(*c.LogFileDir, LiveLogName)
	}
	if c.LogArchiveDir != nil && *c.LogArchiveDir != "" {
		archive = filepath.Join(*c.LogArchiveDir, archiveName)
	}
	return liveLog, archive
}

// LogsToStdout reports whether algod is configured to write to stdout instead of
// node.log, which it does when LogSizeLimit is explicitly zero.
func (c *Config) LogsToStdout() bool {
	return c != nil && c.LogSizeLimit != nil && *c.LogSizeLimit == 0
}

// IsEqual reports whether two configs hold the same values in every field,
// with a field left out of config.json equal only to one left out of the other.
//
// Every field of Config is compared. A field added above has to be added here
// and to MergeAlgodConfigs, or two configs that differ in it compare equal and
// a change to it is dropped on the way to being written.
func (c Config) IsEqual(conf Config) bool {
	return sameOption(c.EnableP2P, conf.EnableP2P) &&
		sameOption(c.EnableP2PHybridMode, conf.EnableP2PHybridMode) &&
		sameOption(c.BaseLoggerDebugLevel, conf.BaseLoggerDebugLevel) &&
		sameOption(c.LogSizeLimit, conf.LogSizeLimit) &&
		sameOption(c.HotDataDir, conf.HotDataDir) &&
		sameOption(c.ColdDataDir, conf.ColdDataDir) &&
		sameOption(c.LogFileDir, conf.LogFileDir) &&
		sameOption(c.LogArchiveDir, conf.LogArchiveDir) &&
		sameOption(c.LogArchiveName, conf.LogArchiveName)
}

// MergeAlgodConfigs merges two Config objects, with every field 'b' sets
// overriding the one in 'a'. A field 'b' leaves unset keeps a's value, which is
// what makes it safe to build the override from the flags the user actually
// passed.
func MergeAlgodConfigs(a Config, b Config) Config {
	// Field by field onto a copy of a, rather than onto a fresh Config: a field
	// added above and forgotten here then keeps a's value instead of being
	// silently cleared out of the merge.
	merged := a

	merged.EnableP2P = override(a.EnableP2P, b.EnableP2P)
	merged.EnableP2PHybridMode = override(a.EnableP2PHybridMode, b.EnableP2PHybridMode)
	merged.BaseLoggerDebugLevel = override(a.BaseLoggerDebugLevel, b.BaseLoggerDebugLevel)
	merged.LogSizeLimit = override(a.LogSizeLimit, b.LogSizeLimit)
	merged.HotDataDir = override(a.HotDataDir, b.HotDataDir)
	merged.ColdDataDir = override(a.ColdDataDir, b.ColdDataDir)
	merged.LogFileDir = override(a.LogFileDir, b.LogFileDir)
	merged.LogArchiveDir = override(a.LogArchiveDir, b.LogArchiveDir)
	merged.LogArchiveName = override(a.LogArchiveName, b.LogArchiveName)

	return merged
}

// sameOption reports whether two optional fields say the same thing. A field
// that is absent is not the same as one set to the zero value: algod's own
// defaults are not all zero, so "not in config.json" means "whatever algod
// defaults to" and nothing else.
func sameOption[T comparable](a, b *T) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	return a == nil || *a == *b
}

// override returns the value a merge should take for one field: b's when it
// sets one, and a's otherwise.
func override[T comparable](a, b *T) *T {
	if b != nil {
		return b
	}
	return a
}
