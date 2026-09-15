package statewalker

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func Test_ParseSpeed(t *testing.T) {
	if speed, err := ParseSpeed("fast"); err != nil || speed != SpeedFast {
		t.Errorf("expected fast to parse, got %s, %v", speed, err)
	}
	if speed, err := ParseSpeed("real"); err != nil || speed != SpeedReal {
		t.Errorf("expected real to parse, got %s, %v", speed, err)
	}
	if _, err := ParseSpeed("warp"); err == nil {
		t.Error("expected an unknown profile to be rejected")
	}
}

func Test_SpeedProfile_ConsensusOverrides(t *testing.T) {
	if overrides := SpeedReal.ConsensusOverrides(); overrides != nil {
		t.Errorf("expected the real profile to change nothing, got %v", overrides)
	}
	overrides := SpeedFast.ConsensusOverrides()
	for _, key := range []string{"AgreementFilterTimeoutPeriod0", "AgreementFilterTimeout"} {
		raw, ok := overrides[key]
		if !ok {
			t.Errorf("expected the fast profile to override %s", key)
			continue
		}
		var ns int64
		if err := json.Unmarshal(raw, &ns); err != nil || ns <= 0 {
			t.Errorf("expected %s to be a positive duration in ns, got %s (%v)", key, raw, err)
		}
	}
	// Safety-relevant parameters must never be touched
	for _, key := range []string{"SeedLookback", "SeedRefreshInterval", "UpgradeVoteRounds"} {
		if _, ok := overrides[key]; ok {
			t.Errorf("the fast profile must not override %s", key)
		}
	}
}

func Test_ApplySpeed(t *testing.T) {
	dir := t.TempDir()
	nodeA := filepath.Join(dir, "Primary")
	nodeB := filepath.Join(dir, "Node2")
	for _, node := range []string{nodeA, nodeB} {
		if err := os.MkdirAll(node, 0755); err != nil {
			t.Fatal(err)
		}
	}
	genesis := `{"network":"tuinet","id":"v1","proto":"test-proto"}`
	if err := os.WriteFile(filepath.Join(nodeA, "genesis.json"), []byte(genesis), 0644); err != nil {
		t.Fatal(err)
	}

	// Stub the goal protocols call
	originalRunner := runner
	defer func() { runner = originalRunner }()
	runner = func(name string, args ...string) (string, error) {
		return `{"test-proto":{"AgreementFilterTimeout":4000000000,"AgreementFilterTimeoutPeriod0":4000000000,"UpgradeVoteRounds":10000}}`, nil
	}

	// The real profile must not write any consensus override
	if err := ApplySpeed(SpeedReal, []string{nodeA, nodeB}); err != nil {
		t.Fatalf("expected the real profile to be a no-op, got: %s", err)
	}
	if _, err := os.Stat(filepath.Join(nodeA, "consensus.json")); !os.IsNotExist(err) {
		t.Error("expected no consensus.json for the real profile")
	}

	// The fast profile writes the merged table into every node dir
	if err := ApplySpeed(SpeedFast, []string{nodeA, nodeB}); err != nil {
		t.Fatalf("ApplySpeed failed: %s", err)
	}
	for _, node := range []string{nodeA, nodeB} {
		raw, err := os.ReadFile(filepath.Join(node, "consensus.json"))
		if err != nil {
			t.Fatalf("expected consensus.json in %s: %s", node, err)
		}
		var consensus map[string]map[string]json.Number
		if err := json.Unmarshal(raw, &consensus); err != nil {
			t.Fatalf("failed to parse consensus.json: %s", err)
		}
		params, ok := consensus["test-proto"]
		if !ok {
			t.Fatal("expected the genesis protocol entry in consensus.json")
		}
		if params["AgreementFilterTimeoutPeriod0"].String() != "1000000000" {
			t.Errorf("expected the period-0 filter timeout to be overridden, got %s", params["AgreementFilterTimeoutPeriod0"])
		}
		if params["AgreementFilterTimeout"].String() != "2000000000" {
			t.Errorf("expected the filter timeout to be overridden, got %s", params["AgreementFilterTimeout"])
		}
		// Untouched fields of the cloned protocol must be preserved
		if params["UpgradeVoteRounds"].String() != "10000" {
			t.Errorf("expected untouched params to be preserved, got %s", params["UpgradeVoteRounds"])
		}
	}
}

func Test_ApplySpeed_UnknownProto(t *testing.T) {
	dir := t.TempDir()
	node := filepath.Join(dir, "Primary")
	if err := os.MkdirAll(node, 0755); err != nil {
		t.Fatal(err)
	}
	genesis := `{"network":"tuinet","id":"v1","proto":"missing-proto"}`
	if err := os.WriteFile(filepath.Join(node, "genesis.json"), []byte(genesis), 0644); err != nil {
		t.Fatal(err)
	}

	originalRunner := runner
	defer func() { runner = originalRunner }()
	runner = func(name string, args ...string) (string, error) {
		return `{"test-proto":{}}`, nil
	}

	if err := ApplySpeed(SpeedFast, []string{node}); err == nil {
		t.Error("expected an error when the genesis protocol is not in the table")
	}
}
