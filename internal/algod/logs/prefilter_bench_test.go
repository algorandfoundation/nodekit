package logs

import (
	"fmt"
	"testing"
	"time"
)

// benchCorpus mimics the level distribution of a real node.log at algod's
// default logging floor: almost all of it is info that the default view throws
// away, which is the case the prescreen exists for.
func benchCorpus(n int) [][]byte {
	base := time.Date(2026, 9, 9, 14, 0, 0, 0, time.UTC)
	out := make([][]byte, 0, n)
	for i := 0; i < n; i++ {
		ts := base.Add(time.Duration(i) * 300 * time.Millisecond).Format(time.RFC3339Nano)
		level, msg := "info", "Sync round set to 48291043 for services"
		switch {
		case i%100 == 0:
			level, msg = "warning", "fetchRound could not acquire block"
		case i%250 == 0:
			level, msg = "error", "peer 1.2.3.4:4160 error: connection reset by peer"
		}
		out = append(out, fmt.Appendf(nil,
			`{"file":"service.go","function":"github.com/algorand/go-algorand/agreement.(*Service).mainLoop","level":%q,"line":%d,"msg":%q,"time":%q}`,
			level, 200+i%800, msg, ts))
	}
	return out
}

func benchBytes(corpus [][]byte) int64 {
	var n int64
	for _, line := range corpus {
		n += int64(len(line)) + 1
	}
	return n
}

// The pair to compare: what the scan costs per line with and without the
// prescreen in front of ParseLine.
func BenchmarkFilterWithoutPrescreen(b *testing.B) {
	corpus := benchCorpus(20000)
	f := Filter{MinLevel: LevelWarn}
	b.SetBytes(benchBytes(corpus))
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		for _, line := range corpus {
			if e := ParseLine(line); f.Keep(e) {
				runtimeKeepAlive(e)
			}
		}
	}
}

func BenchmarkFilterWithPrescreen(b *testing.B) {
	corpus := benchCorpus(20000)
	f := Filter{MinLevel: LevelWarn}
	pre := newPrescreen(f)
	b.SetBytes(benchBytes(corpus))
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		for _, line := range corpus {
			if pre.rejects(line) {
				continue
			}
			if e := ParseLine(line); f.Keep(e) {
				runtimeKeepAlive(e)
			}
		}
	}
}

func BenchmarkFilterWithTextPrescreen(b *testing.B) {
	corpus := benchCorpus(20000)
	f := Filter{MinLevel: LevelTrace, Text: "connection reset"}
	pre := newPrescreen(f)
	b.SetBytes(benchBytes(corpus))
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		for _, line := range corpus {
			if pre.rejects(line) {
				continue
			}
			if e := ParseLine(line); f.Keep(e) {
				runtimeKeepAlive(e)
			}
		}
	}
}

var sink Entry

func runtimeKeepAlive(e Entry) { sink = e }
