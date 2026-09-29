package system

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

const reexecTestHelperEnv = "NODEKIT_REEXEC_TEST_HELPER"

func TestReexecCommand(t *testing.T) {
	args := []string{"upgrade", "--no-incentives"}
	cmd := reexecCommand("/usr/local/bin/nodekit", args, []string{"PATH=/usr/bin"})

	if want := []string{"/usr/local/bin/nodekit", "upgrade", "--no-incentives"}; !reflect.DeepEqual(cmd.Args, want) {
		t.Errorf("command args = %q, want %q", cmd.Args, want)
	}
	if cmd.Stdin != os.Stdin || cmd.Stdout != os.Stdout || cmd.Stderr != os.Stderr {
		t.Error("re-executed command does not inherit terminal streams")
	}
	if want := []string{"PATH=/usr/bin", reexecEnv + "=1"}; !reflect.DeepEqual(cmd.Env, want) {
		t.Errorf("command environment = %q, want %q", cmd.Env, want)
	}
}

func TestReexecCommandStartFailure(t *testing.T) {
	cmd := reexecCommand(filepath.Join(t.TempDir(), "missing-nodekit"), nil, nil)
	if err := cmd.Start(); err == nil {
		t.Error("starting a missing executable succeeded")
	}
}

func TestReexecWaitsForChild(t *testing.T) {
	if os.Getenv(reexecTestHelperEnv) == "1" {
		os.Exit(42)
	}

	t.Setenv(reexecTestHelperEnv, "1")
	err := Reexec(os.Args[0], []string{"-test.run=^TestReexecWaitsForChild$"})
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("Reexec() error = %v, want child exit error", err)
	}
	if exitErr.ExitCode() != 42 {
		t.Errorf("child exit code = %d, want 42", exitErr.ExitCode())
	}
}

func TestIsReexec(t *testing.T) {
	t.Setenv(reexecEnv, "1")
	if !IsReexec() {
		t.Error("IsReexec() = false, want true when re-exec marker is set")
	}

	t.Setenv(reexecEnv, "")
	if IsReexec() {
		t.Error("IsReexec() = true, want false when re-exec marker is absent")
	}
}
