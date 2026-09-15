package cmd

import (
	"context"
	"fmt"
	"strings"
	"time"

	statewalker "github.com/algorandfoundation/nodekit/tools/statewalker/internal"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// partkeyAddress optionally pins partkey commands to a specific Primary wallet address.
var partkeyAddress string

// expiring holds the validity durations for `partkeys seed`.
var expiring []time.Duration

// registerOnline makes `partkeys seed` register the account online with the last generated key.
var registerOnline bool

// partkeysCmd is the parent for participation key scenario commands.
var partkeysCmd = &cobra.Command{
	Use:   "partkeys",
	Short: "Drive participation key states on the Primary node",
}

var partkeysSeedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Generate participation keys with configurable validity windows",
	Long: `Generates one participation key per --expiring duration on the Primary
node's online test wallet. Durations are converted to rounds using the network's
measured average block time, matching the expiration math of the nodekit TUI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dataDir, err := requirePrivate()
		if err != nil {
			return err
		}
		ctx := context.Background()
		client, err := getClient(ctx, dataDir)
		if err != nil {
			return err
		}
		account, err := resolvePrimaryAccount(dataDir, partkeyAddress, true)
		if err != nil {
			return err
		}
		roundTime, err := statewalker.MeasureRoundTime(ctx, client)
		if err != nil {
			return err
		}
		log.Info("Measured average round time", "roundTime", roundTime)

		for _, duration := range expiring {
			status, _, err := statewalker.CurrentStatus(ctx, client)
			if err != nil {
				return err
			}
			rounds := statewalker.DurationToRounds(duration, roundTime)
			first := status.LastRound
			last := first + rounds
			log.Info("Generating participation key",
				"address", account.Address,
				"expiring", duration,
				"firstValid", first,
				"lastValid", last,
			)
			if _, err := statewalker.AddPartKey(dataDir, account.Address, first, last); err != nil {
				return err
			}
			fmt.Printf("key valid for ~%s (rounds %d-%d), expected expiration ~%s\n",
				duration, first, last, time.Now().Add(duration).Format(time.RFC1123))
		}

		if registerOnline {
			log.Info("Registering account online", "address", account.Address)
			if _, err := statewalker.SetOnlineStatus(dataDir, account.Address, true); err != nil {
				return err
			}
		}
		return nil
	},
}

var partkeysOnlineCmd = &cobra.Command{
	Use:   "online",
	Short: "Register a Primary wallet account online",
	RunE: func(cmd *cobra.Command, args []string) error {
		return changeOnlineStatus(true)
	},
}

var partkeysOfflineCmd = &cobra.Command{
	Use:   "offline",
	Short: "Register a Primary wallet account offline (anchor stake keeps the chain alive)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return changeOnlineStatus(false)
	},
}

// changeOnlineStatus flips the online status of a Primary wallet account and
// verifies the chain still advances afterwards (the anchor holds quorum).
func changeOnlineStatus(online bool) error {
	dataDir, err := requirePrivate()
	if err != nil {
		return err
	}
	ctx := context.Background()
	client, err := getClient(ctx, dataDir)
	if err != nil {
		return err
	}
	account, err := resolvePrimaryAccount(dataDir, partkeyAddress, !online)
	if err != nil {
		return err
	}
	log.Info("Changing online status", "address", account.Address, "online", online)
	out, err := statewalker.SetOnlineStatus(dataDir, account.Address, online)
	if err != nil && online && strings.Contains(err.Error(), "couldn't find a participation key") {
		// No valid key installed (e.g. after an expiry scenario); seed one and retry
		status, _, statusErr := statewalker.CurrentStatus(ctx, client)
		if statusErr != nil {
			return statusErr
		}
		log.Info("No valid participation key found; generating one", "address", account.Address)
		if _, err = statewalker.AddPartKey(dataDir, account.Address, status.LastRound, status.LastRound+1_000_000); err != nil {
			return err
		}
		out, err = statewalker.SetOnlineStatus(dataDir, account.Address, online)
	}
	if err != nil {
		return err
	}
	fmt.Print(out)
	log.Info("Verifying rounds still advance")
	if err := statewalker.WaitForRoundsToAdvance(ctx, client, 2, time.Minute); err != nil {
		return fmt.Errorf("chain stalled after status change; anchor stake may be compromised: %w", err)
	}
	log.Info("Chain is still advancing")
	return nil
}

var partkeysZombieCmd = &cobra.Command{
	Use:   "zombie",
	Short: "Install a participation key for an account that is not registered with it",
	Long: `Reproduces the IS #41 hazard: installs a participation key on the Primary
node for the offline funded wallet without registering the account online.
The key shows up in the TUI's key list although the account never uses it.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dataDir, err := requirePrivate()
		if err != nil {
			return err
		}
		ctx := context.Background()
		client, err := getClient(ctx, dataDir)
		if err != nil {
			return err
		}
		account, err := resolvePrimaryAccount(dataDir, partkeyAddress, false)
		if err != nil {
			return err
		}
		status, _, err := statewalker.CurrentStatus(ctx, client)
		if err != nil {
			return err
		}
		first := status.LastRound
		last := first + 1_000_000
		log.Info("Installing zombie participation key",
			"address", account.Address,
			"firstValid", first,
			"lastValid", last,
		)
		if _, err := statewalker.AddPartKey(dataDir, account.Address, first, last); err != nil {
			return err
		}
		fmt.Printf("zombie key installed for offline account %s (rounds %d-%d); the account is NOT registered with it\n",
			account.Address, first, last)
		return nil
	},
}

var partkeysListCmd = &cobra.Command{
	Use:   "list",
	Short: "List participation keys installed on the Primary node",
	RunE: func(cmd *cobra.Command, args []string) error {
		dataDir, err := requirePrivate()
		if err != nil {
			return err
		}
		out, err := statewalker.PartKeyInfo(dataDir)
		if err != nil {
			return err
		}
		fmt.Print(out)
		return nil
	},
}

func init() {
	partkeysCmd.PersistentFlags().StringVarP(&partkeyAddress, "address", "a", "", "Primary wallet address to target (defaults to a suitable one)")
	partkeysSeedCmd.Flags().DurationSliceVar(&expiring, "expiring", []time.Duration{72 * time.Hour, 720 * time.Hour}, "validity durations of the generated keys (repeatable)")
	partkeysSeedCmd.Flags().BoolVar(&registerOnline, "online", false, "register the account online after seeding")
	partkeysCmd.AddCommand(partkeysSeedCmd)
	partkeysCmd.AddCommand(partkeysOnlineCmd)
	partkeysCmd.AddCommand(partkeysOfflineCmd)
	partkeysCmd.AddCommand(partkeysZombieCmd)
	partkeysCmd.AddCommand(partkeysListCmd)
	rootCmd.AddCommand(partkeysCmd)
}
