package config

// Config represents the configuration settings for algod, including enabling P2PHybrid
type Config struct {
	EnableP2P           *bool `json:"EnableP2P,omitempty"`
	EnableP2PHybridMode *bool `json:"EnableP2PHybridMode,omitempty"`
}

// IsEqual compares two Config objects and returns true if all their fields have the same values, otherwise false.
func (c Config) IsEqual(conf Config) bool {
	// Check EnableP2P
	if (c.EnableP2P == nil) != (conf.EnableP2P == nil) {
		return false
	}
	if c.EnableP2P != nil && *c.EnableP2P != *conf.EnableP2P {
		return false
	}

	// Check EnableP2PHybridMode
	if (c.EnableP2PHybridMode == nil) != (conf.EnableP2PHybridMode == nil) {
		return false
	}
	if c.EnableP2PHybridMode != nil && *c.EnableP2PHybridMode != *conf.EnableP2PHybridMode {
		return false
	}

	return true
}

// MergeAlgodConfigs merges two Config objects, with non-zero and non-default fields in 'b' overriding those in 'a'.
func MergeAlgodConfigs(a Config, b Config) Config {
	merged := a

	if b.EnableP2P != nil {
		if a.EnableP2P == nil || *b.EnableP2P != *a.EnableP2P {
			merged.EnableP2P = b.EnableP2P
		}
	}

	if b.EnableP2PHybridMode != nil {
		if a.EnableP2PHybridMode == nil || *b.EnableP2PHybridMode != *a.EnableP2PHybridMode {
			merged.EnableP2PHybridMode = b.EnableP2PHybridMode
		}
	}

	return merged
}
