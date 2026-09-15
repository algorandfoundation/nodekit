package system

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

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
