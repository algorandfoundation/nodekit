package utils

import (
	"testing"
)

func Test_Nicknames(t *testing.T) {
	// Isolate the settings file to a temporary HOME so we don't touch the
	// developer's real ~/.nodekit.json.
	t.Setenv("HOME", t.TempDir())

	const addr = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	// No settings file yet -> empty, non-nil map.
	names, err := GetNicknames()
	if err != nil {
		t.Fatalf("GetNicknames returned error: %v", err)
	}
	if names == nil {
		t.Fatal("GetNicknames returned a nil map")
	}
	if len(names) != 0 {
		t.Fatalf("expected no nicknames, got %v", names)
	}

	// Set and read back.
	if err := SetNickname(addr, "my-node"); err != nil {
		t.Fatalf("SetNickname returned error: %v", err)
	}
	names, err = GetNicknames()
	if err != nil {
		t.Fatalf("GetNicknames returned error: %v", err)
	}
	if names[addr] != "my-node" {
		t.Fatalf("expected nickname %q, got %q", "my-node", names[addr])
	}

	// Whitespace is trimmed.
	if err := SetNickname(addr, "  spaced  "); err != nil {
		t.Fatalf("SetNickname returned error: %v", err)
	}
	names, _ = GetNicknames()
	if names[addr] != "spaced" {
		t.Fatalf("expected trimmed nickname %q, got %q", "spaced", names[addr])
	}

	// An empty (or whitespace-only) nickname clears the entry.
	if err := SetNickname(addr, "   "); err != nil {
		t.Fatalf("SetNickname returned error: %v", err)
	}
	names, _ = GetNicknames()
	if _, ok := names[addr]; ok {
		t.Fatalf("expected nickname to be cleared, got %q", names[addr])
	}
}
