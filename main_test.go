package main

import (
	"testing"

	"github.com/algorandfoundation/nodekit/api"
)

// Test_UserAgentVersion guards the hand-off in init: whatever the linker
// stamps into this package is what outbound requests report. A test binary is
// not linked as package main, so -X main.version does not reach it and the
// value here is always the "dev" default; the point is that the two agree.
func Test_UserAgentVersion(t *testing.T) {
	if got := api.UserAgent(); got != "Nodekit v"+version {
		t.Fatalf("expected Nodekit v%s, got %s", version, got)
	}
}
