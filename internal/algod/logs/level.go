package logs

import (
	"fmt"
	"strings"
)

// Level represents the severity of a log entry.
//
// Values ascend with severity so that filtering is a single >= comparison.
// Note this is the reverse of logrus' own numbering (Panic=0 .. Trace=6) and of
// go-algorand's logging.Level (Panic=0 .. Debug=5).
type Level int8

const (
	// LevelUnknown is used for lines that could not be parsed, or whose "level"
	// field was missing or unrecognised. These are never filtered out by
	// severity alone, see Filter.Keep.
	LevelUnknown Level = iota
	LevelTrace
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
	LevelPanic
)

// LevelNames lists the accepted --level values, from least to most severe.
var LevelNames = []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}

// String returns the lowercase name of the level, as algod would spell it in
// node.log. Note that LevelWarn renders as "warn" here for display purposes even
// though logrus writes "warning" on the wire.
func (l Level) String() string {
	switch l {
	case LevelTrace:
		return "trace"
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelWarn:
		return "warn"
	case LevelError:
		return "error"
	case LevelFatal:
		return "fatal"
	case LevelPanic:
		return "panic"
	default:
		return "unknown"
	}
}

// ParseLevel converts a level name to a Level. It is case-insensitive and
// accepts the spellings logrus actually writes to node.log.
//
// The "warning" alias is essential: logrus marshals WarnLevel as "warning", not
// "warn", so matching only "warn" silently matches nothing at all.
func ParseLevel(s string) (Level, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "trace":
		return LevelTrace, nil
	case "debug":
		return LevelDebug, nil
	case "info":
		return LevelInfo, nil
	case "warn", "warning":
		return LevelWarn, nil
	case "error", "err":
		return LevelError, nil
	case "fatal":
		return LevelFatal, nil
	case "panic":
		return LevelPanic, nil
	default:
		return LevelUnknown, fmt.Errorf("invalid level %q: expected one of %s", s, strings.Join(LevelNames, ", "))
	}
}

// AlgodLevel is the value of config.json's BaseLoggerDebugLevel, which algod
// passes straight to logging.Level. It decides what algod ever writes to
// node.log, as opposed to what we choose to display.
//
// The numbering is go-algorand's logging.Level: Panic=0 .. Debug=5. There is no
// Trace, so a node can never emit trace entries.
type AlgodLevel uint32

// Level converts algod's configured logging floor to the equivalent Level.
func (a AlgodLevel) Level() Level {
	switch a {
	case 0:
		return LevelPanic
	case 1:
		return LevelFatal
	case 2:
		return LevelError
	case 3:
		return LevelWarn
	case 4:
		return LevelInfo
	default:
		// 5 is Debug; anything higher is not meaningful to algod but is at
		// least as verbose, so treat it as Debug too.
		return LevelDebug
	}
}
