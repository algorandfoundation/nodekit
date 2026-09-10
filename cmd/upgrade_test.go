package cmd

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/algorandfoundation/nodekit/api"
	"github.com/charmbracelet/log"
	"github.com/stretchr/testify/require"
)

func TestUpgradeLogsSuccessMessages(t *testing.T) {
	originalNeedsUpgrade := NeedsUpgrade
	originalNodeKitUpgrade := nodeKitUpgrade
	originalAlgodUpgrade := algodUpgrade
	originalAlgodIsRunning := algodIsRunning
	originalAlgodStart := algodStart
	originalUpgradeSleep := upgradeSleep

	t.Cleanup(func() {
		NeedsUpgrade = originalNeedsUpgrade
		nodeKitUpgrade = originalNodeKitUpgrade
		algodUpgrade = originalAlgodUpgrade
		algodIsRunning = originalAlgodIsRunning
		algodStart = originalAlgodStart
		upgradeSleep = originalUpgradeSleep
		log.SetOutput(os.Stderr)
	})

	var output bytes.Buffer
	log.SetOutput(&output)
	NeedsUpgrade = true
	nodeKitUpgrade = func(api.HttpPkgInterface) error { return nil }
	algodUpgrade = func() error { return nil }
	algodIsRunning = func(string) bool { return false }
	algodStarted := false
	algodStart = func() error {
		algodStarted = true
		return nil
	}
	upgradeSleep = func(time.Duration) {}

	upgradeCmd.Run(upgradeCmd, nil)

	require.Contains(t, output.String(), NodeKitUpgradeSuccessMsg)
	require.Contains(t, output.String(), AlgodUpgradeSuccessMsg)
	require.True(t, algodStarted)
}
