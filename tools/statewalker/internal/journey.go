package statewalker

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/charmbracelet/log"
)

const (
	// JourneyImage is the docker image tag built for the journey container.
	JourneyImage = "statewalker-journey"

	// JourneyContainer is the name of the throwaway journey container.
	JourneyContainer = "statewalker-journey"

	// JourneyDataDir is the algorand data directory inside the container.
	JourneyDataDir = "/var/lib/algorand"

	// JourneyRestPort is the container-side REST port published to the host.
	JourneyRestPort = "8080"

	// JourneyUser is the non-root end-user account the journey runs as. It has
	// passwordless sudo (see journey/Dockerfile) so nodekit's `sudo apt-get`
	// calls never block on a password prompt.
	JourneyUser = "algo"

	// JourneyHome is the journey user's home directory, the working directory
	// of every user-flow command (where install.sh drops the binary).
	JourneyHome = "/home/algo"

	// JourneyExecInactivity is how long an exec may stay completely silent
	// before the watchdog kills it and names the hung command. Anything
	// waiting on interactive input (a sudo password, a debconf question)
	// produces no output and no CPU, so silence is the reliable signal.
	JourneyExecInactivity = 4 * time.Minute
)

// journeyUserEnv keeps every package operation prompt-free inside the journey.
var journeyUserEnv = []string{
	"DEBIAN_FRONTEND=noninteractive",
	"APT_LISTCHANGES_FRONTEND=none",
}

// Backend abstracts the journey environment (exec + lifecycle) so scenario
// phases stay agnostic of where the end-user flow runs. The Docker/Debian
// container is the first implementation; a macOS backend (brew/launchd on a
// mac runner) can be added later without touching the phases.
type Backend interface {
	// Provision builds and starts the environment, leaving it ready for Exec.
	Provision() error

	// Exists reports whether the environment is currently provisioned.
	Exists() bool

	// Exec runs a script as the end user (non-root, passwordless sudo) with
	// optional extra KEY=VALUE environment variables.
	Exec(script string, env ...string) (string, error)

	// ExecRoot runs orchestration scaffolding (genesis injection, service
	// rewiring) with full privileges.
	ExecRoot(script string) (string, error)

	// CopyTo copies a host file into the environment.
	CopyTo(src string, dst string) error

	// WriteFile writes content to a path inside the environment.
	WriteFile(path string, content string) error

	// ReadFile reads a file from inside the environment.
	ReadFile(path string) (string, error)

	// Endpoint returns the host address reaching the node's REST port.
	Endpoint(port string) (string, error)

	// Logs returns the tail of the environment's own stdout/stderr.
	Logs(tail int) string

	// Remove tears the environment down; a missing one is not an error.
	Remove() error
}

// DockerBackend runs the journey in a throwaway Debian/systemd container.
type DockerBackend struct {
	Container  string
	Image      string
	Dockerfile []byte
}

// Provision verifies docker, removes any leftover container, builds the image
// and starts a fresh container, waiting for systemd to become ready.
func (d DockerBackend) Provision() error {
	if err := PreflightDocker(); err != nil {
		return err
	}
	if d.Exists() {
		log.Info("Removing leftover journey container", "container", d.Container)
		if err := DockerRemove(d.Container); err != nil {
			return err
		}
	}
	log.Info("Building the journey image", "image", d.Image)
	if err := DockerBuild(d.Image, d.Dockerfile); err != nil {
		return err
	}
	log.Info("Starting the journey container", "container", d.Container)
	if err := DockerRunJourney(d.Container, d.Image); err != nil {
		return err
	}
	if _, err := d.ExecRoot("timeout 60 bash -c 'until systemctl is-system-running &>/dev/null || [ \"$(systemctl is-system-running 2>/dev/null)\" = degraded ]; do sleep 1; done'"); err != nil {
		return fmt.Errorf("systemd did not become ready in the container: %w", err)
	}
	return nil
}

// Exists reports whether the journey container is present.
func (d DockerBackend) Exists() bool {
	return ContainerExists(d.Container)
}

// Exec runs a script as the journey user in its home directory, with the
// noninteractive frontends set and the inactivity watchdog armed.
func (d DockerBackend) Exec(script string, env ...string) (string, error) {
	args := []string{"exec", "-u", JourneyUser, "-w", JourneyHome}
	for _, kv := range append(append([]string{}, journeyUserEnv...), env...) {
		args = append(args, "-e", kv)
	}
	args = append(args, d.Container, "bash", "-lc", script)
	return watchdogExec(script, JourneyExecInactivity, "docker", args...)
}

// ExecRoot runs orchestration scaffolding as root, still watchdog-guarded.
func (d DockerBackend) ExecRoot(script string) (string, error) {
	return watchdogExec(script, JourneyExecInactivity, "docker", "exec", d.Container, "bash", "-lc", script)
}

// CopyTo copies a host file into the container.
func (d DockerBackend) CopyTo(src string, dst string) error {
	return DockerCopyTo(d.Container, src, dst)
}

// WriteFile writes content to a path inside the container.
func (d DockerBackend) WriteFile(path string, content string) error {
	return DockerWriteFile(d.Container, path, content)
}

// ReadFile reads a file from inside the container.
func (d DockerBackend) ReadFile(path string) (string, error) {
	return DockerExec(d.Container, "cat "+path)
}

// Endpoint returns the host endpoint publishing the given container port.
func (d DockerBackend) Endpoint(port string) (string, error) {
	return DockerPort(d.Container, port)
}

// Logs returns the tail of the container's stdout/stderr.
func (d DockerBackend) Logs(tail int) string {
	return DockerLogs(d.Container, tail)
}

// Remove force-removes the container.
func (d DockerBackend) Remove() error {
	return DockerRemove(d.Container)
}

// watchdogExec runs a command, killing it if it produces no output for the
// given inactivity window. A killed command fails immediately with the exact
// script named, instead of stalling at 0% CPU behind an interactive prompt.
func watchdogExec(label string, inactivity time.Duration, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	pr, pw, err := os.Pipe()
	if err != nil {
		return "", err
	}
	cmd.Stdout = pw
	cmd.Stderr = pw
	if err := cmd.Start(); err != nil {
		pw.Close()
		pr.Close()
		return "", err
	}
	pw.Close()

	var buf bytes.Buffer
	var lastActivity atomic.Int64
	lastActivity.Store(time.Now().UnixNano())
	readerDone := make(chan struct{})
	go func() {
		defer close(readerDone)
		chunk := make([]byte, 4096)
		for {
			n, err := pr.Read(chunk)
			if n > 0 {
				buf.Write(chunk[:n])
				lastActivity.Store(time.Now().UnixNano())
			}
			if err != nil {
				return
			}
		}
	}()

	var killed atomic.Bool
	watchdogDone := make(chan struct{})
	go func() {
		ticker := time.NewTicker(inactivity / 10)
		defer ticker.Stop()
		for {
			select {
			case <-watchdogDone:
				return
			case <-ticker.C:
				if time.Since(time.Unix(0, lastActivity.Load())) > inactivity {
					killed.Store(true)
					_ = cmd.Process.Kill()
					return
				}
			}
		}
	}()

	waitErr := cmd.Wait()
	close(watchdogDone)
	<-readerDone
	pr.Close()

	out := buf.String()
	if killed.Load() {
		return out, fmt.Errorf("command produced no output for %s and was killed; it was likely waiting for interactive input (a prompt): %s\n%s", inactivity, label, out)
	}
	if waitErr != nil {
		return out, fmt.Errorf("exec failed: %w\n%s", waitErr, out)
	}
	return out, nil
}

// PreflightDocker verifies the docker CLI is installed and the daemon answers.
func PreflightDocker() error {
	if _, err := lookPath("docker"); err != nil {
		return fmt.Errorf("docker is required for the journey scenario: %w", err)
	}
	if out, err := runner("docker", "info", "--format", "{{.ServerVersion}}"); err != nil {
		return fmt.Errorf("docker daemon is not reachable: %w\n%s", err, out)
	}
	return nil
}

// DockerBuild builds the journey image from the embedded Dockerfile,
// materialized into a throwaway build context.
func DockerBuild(image string, dockerfile []byte) error {
	context, err := os.MkdirTemp("", "statewalker-journey-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(context)
	if err := os.WriteFile(filepath.Join(context, "Dockerfile"), dockerfile, 0644); err != nil {
		return err
	}
	out, err := runner("docker", "build", "-t", image, context)
	if err != nil {
		return fmt.Errorf("docker build failed: %w\n%s", err, out)
	}
	return nil
}

// DockerRunJourney starts the systemd journey container detached, publishing
// the algod REST port on an ephemeral localhost port and mapping the docker
// host gateway so the container can dial the host relay.
func DockerRunJourney(name string, image string) error {
	out, err := runner("docker", "run", "-d",
		"--name", name,
		"--privileged",
		"--cgroupns=host",
		"-v", "/sys/fs/cgroup:/sys/fs/cgroup:rw",
		"--tmpfs", "/run",
		"--tmpfs", "/run/lock",
		"--add-host=host.docker.internal:host-gateway",
		"-p", "127.0.0.1:0:"+JourneyRestPort,
		image,
	)
	if err != nil {
		return fmt.Errorf("docker run failed: %w\n%s", err, out)
	}
	return nil
}

// DockerExec runs a bash script inside the container and returns its combined output.
func DockerExec(name string, script string) (string, error) {
	out, err := runner("docker", "exec", name, "bash", "-lc", script)
	if err != nil {
		return out, fmt.Errorf("docker exec failed: %w\n%s", err, out)
	}
	return out, nil
}

// DockerExecEnv runs a bash script inside the container with extra environment
// variables (KEY=VALUE) set.
func DockerExecEnv(name string, env []string, script string) (string, error) {
	args := []string{"exec"}
	for _, kv := range env {
		args = append(args, "-e", kv)
	}
	args = append(args, name, "bash", "-lc", script)
	out, err := runner("docker", args...)
	if err != nil {
		return out, fmt.Errorf("docker exec failed: %w\n%s", err, out)
	}
	return out, nil
}

// DockerCopyTo copies a host file into the container.
func DockerCopyTo(name string, src string, dst string) error {
	out, err := runner("docker", "cp", src, name+":"+dst)
	if err != nil {
		return fmt.Errorf("docker cp failed: %w\n%s", err, out)
	}
	return nil
}

// DockerWriteFile writes content to a path inside the container.
func DockerWriteFile(name string, path string, content string) error {
	tmp, err := os.CreateTemp("", "statewalker-inject-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.WriteString(content); err != nil {
		tmp.Close()
		return err
	}
	tmp.Close()
	return DockerCopyTo(name, tmp.Name(), path)
}

// DockerReadFile reads a file from inside the container.
func DockerReadFile(name string, path string) (string, error) {
	return DockerExec(name, "cat "+path)
}

// DockerPort returns the host endpoint (127.0.0.1:port) publishing the given
// container port.
func DockerPort(name string, containerPort string) (string, error) {
	out, err := runner("docker", "port", name, containerPort+"/tcp")
	if err != nil {
		return "", fmt.Errorf("docker port failed: %w\n%s", err, out)
	}
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "127.0.0.1:") || strings.Contains(line, "0.0.0.0:") {
			return "127.0.0.1:" + line[strings.LastIndex(line, ":")+1:], nil
		}
	}
	return "", fmt.Errorf("no published endpoint found for port %s:\n%s", containerPort, out)
}

// DockerLogs returns the last lines of the container's stdout/stderr.
func DockerLogs(name string, tail int) string {
	out, _ := runner("docker", "logs", "--tail", fmt.Sprintf("%d", tail), name)
	return out
}

// DockerRemove force-removes the container; missing containers are not an error.
func DockerRemove(name string) error {
	out, err := runner("docker", "rm", "-f", name)
	if err != nil && !strings.Contains(out, "No such container") {
		return fmt.Errorf("docker rm failed: %w\n%s", err, out)
	}
	return nil
}

// ContainerExists reports whether a container with the given name exists.
func ContainerExists(name string) bool {
	out, err := runner("docker", "ps", "-aq", "--filter", "name=^"+name+"$")
	return err == nil && strings.TrimSpace(out) != ""
}

// WriteNodeShim writes a minimal data directory (algod.net plus token files)
// so nodekit and the statewalker verifiers can attach to a remote node (e.g.
// the journey container's published REST port) like a local one.
func WriteNodeShim(dir string, endpoint string, token string, adminToken string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	files := map[string]string{
		"algod.net":         endpoint,
		"algod.token":       token,
		"algod.admin.token": adminToken,
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			return err
		}
	}
	return nil
}
