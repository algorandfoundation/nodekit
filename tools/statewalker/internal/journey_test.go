package statewalker

import (
	"strings"
	"testing"
	"time"
)

func Test_watchdogExec_Success(t *testing.T) {
	out, err := watchdogExec("echo hello", time.Second, "bash", "-c", "echo hello")
	if err != nil {
		t.Fatalf("expected success, got: %s", err)
	}
	if !strings.Contains(out, "hello") {
		t.Errorf("expected output to contain hello, got: %q", out)
	}
}

func Test_watchdogExec_Failure(t *testing.T) {
	out, err := watchdogExec("fail", time.Second, "bash", "-c", "echo oops >&2; exit 3")
	if err == nil {
		t.Fatal("expected a failure")
	}
	if !strings.Contains(err.Error(), "exec failed") {
		t.Errorf("expected an exec failure, got: %s", err)
	}
	if !strings.Contains(out, "oops") {
		t.Errorf("expected stderr to be captured, got: %q", out)
	}
}

func Test_watchdogExec_KillsSilentCommand(t *testing.T) {
	// A command reading stdin with no output mimics an interactive prompt
	// (e.g. sudo asking for a password): the watchdog must kill it and name it.
	start := time.Now()
	_, err := watchdogExec("sudo apt-get install", 300*time.Millisecond, "bash", "-c", "sleep 30")
	if err == nil {
		t.Fatal("expected the watchdog to kill the silent command")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("watchdog took too long to fire: %s", elapsed)
	}
	if !strings.Contains(err.Error(), "waiting for interactive input") {
		t.Errorf("expected the prompt hint in the error, got: %s", err)
	}
	if !strings.Contains(err.Error(), "sudo apt-get install") {
		t.Errorf("expected the hung command to be named, got: %s", err)
	}
}

func Test_watchdogExec_ActivityResetsTheClock(t *testing.T) {
	// Steady output keeps the watchdog quiet even when the total runtime
	// exceeds the inactivity window.
	out, err := watchdogExec("chatty", 400*time.Millisecond, "bash", "-c",
		"for i in 1 2 3 4 5; do echo tick $i; sleep 0.2; done")
	if err != nil {
		t.Fatalf("expected the chatty command to survive, got: %s", err)
	}
	if !strings.Contains(out, "tick 5") {
		t.Errorf("expected full output, got: %q", out)
	}
}
