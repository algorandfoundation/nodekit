package logs

import (
	"fmt"
	"strings"
	"time"
)

// Filter decides which entries are shown. The zero Filter keeps everything.
type Filter struct {
	// MinLevel is the inclusive severity floor.
	MinLevel Level

	// Since, when non-zero, drops entries older than this instant.
	Since time.Time

	// Text, when non-empty, keeps only entries whose message contains it.
	//
	// The match is literal and case-sensitive: what the user typed is what is
	// searched for. A log line is full of characters a pattern language would
	// claim (round=4829, 1.2.3.4:4160, [::1]), so taking the search text at
	// face value is both less surprising and, since the same bytes can then be
	// looked for in the undecoded line, considerably faster.
	Text string
}

// Keep reports whether an entry passes the filter.
//
// Checks run cheapest and most selective first. Two deliberate exemptions:
// an entry with no usable timestamp is never dropped by Since, and an entry we
// could not parse at all is kept unless the caller asked for fatal or panic only
// (plain lines in node.log are overwhelmingly panic dumps and startup output,
// which is precisely what a post-mortem needs to see).
//
// The first exemption is bounded by where the entry sits rather than by this
// test: a scan under Since reads only the region of the log at or after the
// bound, so an untimestamped line far below it is never handed here at all.
// See .decisions/4-Node-Logs.md.
func (f Filter) Keep(e Entry) bool {
	if !f.Since.IsZero() && !e.Time.IsZero() && e.Time.Before(f.Since) {
		return false
	}

	if e.Level == LevelUnknown {
		if f.MinLevel > LevelError {
			return false
		}
	} else if e.Level < f.MinLevel {
		return false
	}

	if f.Text != "" && !strings.Contains(e.Message, f.Text) {
		return false
	}

	return true
}

// sinceLayouts are the absolute timestamp formats accepted by ParseSince, tried
// in order. Layouts without a zone are interpreted in the local timezone, which
// is what algod writes.
var sinceLayouts = []string{
	time.RFC3339Nano,
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"2006-01-02",
	"15:04:05",
	"15:04",
}

// timeOnlyLayouts are the subset of sinceLayouts that carry no date and must be
// anchored to the current day.
var timeOnlyLayouts = map[string]bool{"15:04:05": true, "15:04": true}

// ParseSince interprets s as either a duration back from now (15m, 2h, 90s) or
// an absolute timestamp, and returns the resulting lower bound.
//
// now is a parameter so that callers and tests can pin the clock.
func ParseSince(s string, now time.Time) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, nil
	}

	// Durations are tried first. "2h" is a valid duration and not a valid
	// timestamp, and the two sets do not otherwise overlap.
	if d, err := time.ParseDuration(s); err == nil {
		if d < 0 {
			d = -d // "-15m" and "15m" both mean fifteen minutes ago.
		}
		return now.Add(-d), nil
	}

	for _, layout := range sinceLayouts {
		t, err := time.ParseInLocation(layout, s, now.Location())
		if err != nil {
			continue
		}
		if timeOnlyLayouts[layout] {
			// Anchor a bare time-of-day to today.
			t = time.Date(now.Year(), now.Month(), now.Day(),
				t.Hour(), t.Minute(), t.Second(), 0, now.Location())
		}
		return t, nil
	}

	return time.Time{}, fmt.Errorf(
		"invalid --since value %q: expected a duration like 15m or 2h, "+
			"or a time like 2026-09-09, \"2026-09-09 14:30\", or an RFC3339 timestamp", s)
}
