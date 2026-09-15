package statewalker

import (
	"encoding/json"
	"fmt"
)

// SpeedProfile selects the round-time cadence of the private network.
type SpeedProfile string

const (
	// SpeedReal keeps the stock, MainNet-like agreement cadence (~2.8s rounds),
	// used when tape realism matters.
	SpeedReal SpeedProfile = "real"

	// SpeedFast shortens the agreement filter timeouts so rounds complete in
	// roughly a second, shrinking catchup and expiry waits to fit CI budgets.
	SpeedFast SpeedProfile = "fast"
)

// ParseSpeed validates a --speed flag value.
func ParseSpeed(raw string) (SpeedProfile, error) {
	switch SpeedProfile(raw) {
	case SpeedReal, SpeedFast:
		return SpeedProfile(raw), nil
	}
	return "", fmt.Errorf("unknown speed profile %q: use fast | real", raw)
}

// ConsensusOverrides returns the consensus parameters the profile changes, as
// raw JSON values ready to merge into a cloned protocol table. The real
// profile changes nothing. The fast profile only touches agreement timing
// (filter timeouts are pure liveness knobs); it never alters seed, lookback
// or reward parameters that affect safety or ledger state.
func (s SpeedProfile) ConsensusOverrides() map[string]json.RawMessage {
	if s != SpeedFast {
		return nil
	}
	return map[string]json.RawMessage{
		// The period-0 filter timeout is the main determinant of round time.
		"AgreementFilterTimeoutPeriod0": json.RawMessage("1000000000"), // 1s
		// Keep later periods roomier so recovery rounds still converge.
		"AgreementFilterTimeout": json.RawMessage("2000000000"), // 2s
	}
}

// ApplySpeed clones the genesis protocol entry of the local algod protocol
// table, merges the profile's overrides into it and installs the resulting
// consensus.json into every given node data dir. The real profile is a no-op
// so the network runs the stock consensus table.
func ApplySpeed(profile SpeedProfile, nodeDirs []string) error {
	overrides := profile.ConsensusOverrides()
	if len(overrides) == 0 {
		return nil
	}
	if len(nodeDirs) == 0 {
		return fmt.Errorf("no node data dirs given")
	}
	proto, err := GenesisProto(nodeDirs[0])
	if err != nil {
		return err
	}
	protocols, err := Protocols(nodeDirs[0])
	if err != nil {
		return err
	}
	current, ok := protocols[proto]
	if !ok {
		return fmt.Errorf("genesis protocol %s not found in the local algod protocol table", proto)
	}
	merged := make(map[string]json.RawMessage, len(current))
	for key, value := range current {
		merged[key] = value
	}
	for key, value := range overrides {
		merged[key] = value
	}
	consensus := map[string]map[string]json.RawMessage{proto: merged}
	for _, dir := range nodeDirs {
		if err := WriteConsensus(dir, consensus); err != nil {
			return err
		}
	}
	return nil
}
