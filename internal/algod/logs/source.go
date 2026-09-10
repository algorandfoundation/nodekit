package logs

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/algorandfoundation/nodekit/internal/algod/config"
	"github.com/algorandfoundation/nodekit/internal/algod/utils"
)

// Source describes where a node's log lives and what the node was configured to
// write into it.
type Source struct {
	// DataDir is the resolved algod data directory.
	DataDir string

	// Path is the live log location, which is not necessarily inside DataDir:
	// HotDataDir and LogFileDir both move it. See config.Config.LogPaths.
	Path string

	// Archive is where algod rotates the log to. It is a pattern rather than a
	// path when LogArchiveName carries a time template; see ArchiveFiles.
	Archive string

	// Config is the node's config.json, or nil when it could not be read.
	// An unreadable config is normal for an unprivileged user and is never
	// treated as an error.
	Config *config.Config
}

// ResolveSource works out where the log for a data directory should be.
//
// Reading config.json is best effort: on failure the algod defaults are used,
// which is the right answer for the overwhelming majority of nodes.
func ResolveSource(dataDir string) Source {
	src := Source{DataDir: dataDir}

	if cfg, err := utils.GetConfigFromDataDir(dataDir); err == nil {
		src.Config = cfg
	}

	src.Path, src.Archive = src.Config.LogPaths(dataDir)
	return src
}

// LogsToStdout reports whether the node writes to stdout rather than to a log
// file, in which case no log file will ever exist.
func (s Source) LogsToStdout() bool {
	return s.Config.LogsToStdout()
}

// Floor returns the lowest severity this node writes to its log, and whether
// that could be determined from config.json.
func (s Source) Floor() (Level, bool) {
	if s.Config == nil || s.Config.BaseLoggerDebugLevel == nil {
		return LevelUnknown, false
	}
	return AlgodLevel(*s.Config.BaseLoggerDebugLevel).Level(), true
}

// FloorValue returns the raw BaseLoggerDebugLevel, for use in messages that
// quote the config verbatim.
func (s Source) FloorValue() (uint32, bool) {
	if s.Config == nil || s.Config.BaseLoggerDebugLevel == nil {
		return 0, false
	}
	return *s.Config.BaseLoggerDebugLevel, true
}

// HidesWarnings reports whether the node is configured below warn level, which
// makes the default view of the log structurally incomplete.
func (s Source) HidesWarnings() bool {
	floor, ok := s.Floor()
	return ok && floor > LevelWarn
}

// Hides reports whether entries at the requested level are never written by
// this node, which explains an empty result.
func (s Source) Hides(requested Level) bool {
	floor, ok := s.Floor()
	return ok && requested < floor
}

// archiveTemplate matches the Go template actions algod substitutes into
// LogArchiveName, such as node.archive-{{.Year}}{{.Month}}{{.Day}}.log.
var archiveTemplate = regexp.MustCompile(`{{[^}]*}}`)

// LogFiles returns the files that make up this node's log history, newest
// first: the live log, followed by the rotated archives that still exist.
//
// The live log is always first even when it does not exist, so that a caller
// reading it reports the missing file rather than silently falling back to an
// archive.
func (s Source) LogFiles() []string {
	return append([]string{s.Path}, s.ArchiveFiles()...)
}

// PruneSources drops the archives that cannot hold an entry newer than since.
//
// An archive's modification time is an upper bound on the newest entry inside
// it: algod appends nothing to it after rotating it, and compressing it
// afterwards only moves the timestamp later. So an archive last written before
// the bound holds nothing the filter would keep, and can be skipped without
// being opened at all.
//
// This is the only way to skip a compressed archive, which cannot be seeked
// into, and it is what stops a narrow --since from decompressing a history it
// has no use for. The live log is never dropped: it is where the follow offset
// comes from, and the file whose absence the caller has to explain.
//
// The bound holds for the lines inside that carry no timestamp too, panic
// output among them: nothing in the file was written after the file itself
// was.
func PruneSources(sources []string, since time.Time) []string {
	if since.IsZero() || len(sources) < 2 {
		return sources
	}

	// Capped at one so that appending copies rather than writing over the
	// caller's array.
	kept := sources[:1:1]
	for _, path := range sources[1:] {
		info, err := os.Stat(path)
		if err != nil {
			// Not evidence that it is out of range. Keep it and let the read
			// path decide, since it already knows how to skip an archive it
			// cannot open.
			kept = append(kept, path)
			continue
		}
		if info.ModTime().Before(since) {
			continue
		}
		kept = append(kept, path)
	}
	return kept
}

// ArchiveFiles returns the rotated logs that exist, newest first.
//
// algod treats LogArchiveName as a Go template, so a node configured to keep
// dated archives has several of them at once and the resolved Archive is a
// pattern, not a path. Those are found by globbing, and ordered by modification
// time because the names sort in whatever order the template happens to give.
// Compressed archives are included: algod gzips or bzips the rotated file when
// the name says so, and openSource reads those directly.
func (s Source) ArchiveFiles() []string {
	if s.Archive == "" {
		return nil
	}

	candidates := []string{s.Archive}
	if strings.Contains(s.Archive, "{{") {
		matches, err := filepath.Glob(archiveTemplate.ReplaceAllString(s.Archive, "*"))
		if err != nil {
			return nil
		}
		candidates = matches
	}

	type archive struct {
		path string
		info os.FileInfo
	}
	var found []archive
	for _, path := range candidates {
		if path == s.Path {
			continue // a misconfigured node can point both at one file
		}
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		found = append(found, archive{path, info})
	}

	sort.SliceStable(found, func(i, j int) bool {
		return found[i].info.ModTime().After(found[j].info.ModTime())
	})

	paths := make([]string, 0, len(found))
	for _, a := range found {
		paths = append(paths, a.path)
	}
	return paths
}
