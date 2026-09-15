package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	statewalker "github.com/algorandfoundation/nodekit/tools/statewalker/internal"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// trafficInterval is the delay between self-payments in the traffic loop.
var trafficInterval time.Duration

// trafficCmd is the parent for the background payment loop commands, used to
// advance rounds on localnet (which only moves with traffic) or to add load.
var trafficCmd = &cobra.Command{
	Use:   "traffic",
	Short: "Run a background self-payment loop to advance rounds",
}

// trafficPidFile stores the PID of the detached traffic loop.
func trafficPidFile() string {
	return filepath.Join(networkDir, "traffic.pid")
}

// trafficLogFile receives the output of the detached traffic loop.
func trafficLogFile() string {
	return filepath.Join(networkDir, "traffic.log")
}

var trafficStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the self-payment loop in the background",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := resolveMode(); err != nil {
			return err
		}
		if raw, err := os.ReadFile(trafficPidFile()); err == nil {
			if pid, err := strconv.Atoi(strings.TrimSpace(string(raw))); err == nil {
				if process, err := os.FindProcess(pid); err == nil && process.Signal(nil) == nil {
					return fmt.Errorf("traffic loop already running (pid %d); run `statewalker traffic stop` first", pid)
				}
			}
		}
		self, err := os.Executable()
		if err != nil {
			return err
		}
		logFile, err := os.OpenFile(trafficLogFile(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}
		defer logFile.Close()
		loop := exec.Command(self, "traffic", "run",
			"--network-dir", networkDir,
			"--interval", trafficInterval.String(),
		)
		loop.Stdout = logFile
		loop.Stderr = logFile
		if err := loop.Start(); err != nil {
			return err
		}
		if err := os.WriteFile(trafficPidFile(), []byte(strconv.Itoa(loop.Process.Pid)), 0644); err != nil {
			return err
		}
		log.Info("Traffic loop started", "pid", loop.Process.Pid, "interval", trafficInterval, "log", trafficLogFile())
		return loop.Process.Release()
	},
}

var trafficStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the background self-payment loop",
	RunE: func(cmd *cobra.Command, args []string) error {
		raw, err := os.ReadFile(trafficPidFile())
		if err != nil {
			return fmt.Errorf("no traffic loop found: %w", err)
		}
		pid, err := strconv.Atoi(strings.TrimSpace(string(raw)))
		if err != nil {
			return err
		}
		process, err := os.FindProcess(pid)
		if err == nil {
			if err := process.Kill(); err != nil {
				log.Warn("Failed to kill traffic loop; it may have exited already", "pid", pid, "err", err)
			}
		}
		_ = os.Remove(trafficPidFile())
		log.Info("Traffic loop stopped", "pid", pid)
		return nil
	},
}

// trafficRunCmd is the hidden foreground loop spawned by `traffic start`.
var trafficRunCmd = &cobra.Command{
	Use:    "run",
	Hidden: true,
	Short:  "Run the self-payment loop in the foreground",
	RunE: func(cmd *cobra.Command, args []string) error {
		m, err := resolveMode()
		if err != nil {
			return err
		}
		for {
			if err := sendSelfPayment(m); err != nil {
				log.Error("payment failed", "err", err)
			}
			time.Sleep(trafficInterval)
		}
	},
}

// sendSelfPayment sends a zero-cost self-payment from the richest wallet
// account, using local goal for the private network and the containerized goal
// for algokit localnet.
func sendSelfPayment(m statewalker.Mode) error {
	switch m {
	case statewalker.ModePrivate:
		dataDir := statewalker.NodeDataDir(privateDir(), statewalker.PrimaryNodeName)
		accounts, err := statewalker.ListAccounts(dataDir)
		if err != nil {
			return err
		}
		address := richest(accounts)
		out, err := statewalker.SendPayment(dataDir, address, address, 0)
		if err != nil {
			return err
		}
		log.Info("sent", "output", strings.TrimSpace(out))
	case statewalker.ModeLocalnet:
		out, err := statewalker.LocalnetGoal("account", "list")
		if err != nil {
			return err
		}
		address := ""
		for _, line := range strings.Split(out, "\n") {
			fields := strings.Fields(line)
			if len(fields) >= 4 && strings.HasPrefix(fields[0], "[") {
				address = fields[2]
				break
			}
		}
		if address == "" {
			return fmt.Errorf("no funded account found on localnet:\n%s", out)
		}
		sendOut, err := statewalker.LocalnetGoal("clerk", "send", "-f", address, "-t", address, "-a", "0")
		if err != nil {
			return err
		}
		log.Info("sent", "output", strings.TrimSpace(sendOut))
	default:
		return fmt.Errorf("unknown mode: %s", m)
	}
	return nil
}

// richest returns the address with the highest balance.
func richest(accounts []statewalker.WalletAccount) string {
	address := accounts[0].Address
	var best uint64
	for _, account := range accounts {
		if account.MicroAlgos > best {
			best = account.MicroAlgos
			address = account.Address
		}
	}
	return address
}

func init() {
	trafficCmd.PersistentFlags().DurationVar(&trafficInterval, "interval", 2*time.Second, "delay between payments")
	trafficCmd.AddCommand(trafficStartCmd)
	trafficCmd.AddCommand(trafficStopCmd)
	trafficCmd.AddCommand(trafficRunCmd)
	rootCmd.AddCommand(trafficCmd)
}
