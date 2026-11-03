package mac

import "testing"

func TestParseBrewVersion(t *testing.T) {
	tests := []struct {
		name, json, want string
		wantErr          bool
	}{
		{"linked", `{"formulae":[{"linked_keg":"4.1.0","installed":[{"version":"4.0.2"},{"version":"4.1.0"}]}]}`, "4.1.0", false},
		{"single install, not linked", `{"formulae":[{"linked_keg":null,"installed":[{"version":"4.0.2"}]}]}`, "4.0.2", false},
		{"several installs, none linked", `{"formulae":[{"installed":[{"version":"4.0.2"},{"version":"4.1.0"}]}]}`, "", true},
		{"no formula", `{"formulae":[]}`, "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseBrewVersion([]byte(tt.json))
			if (err != nil) != tt.wantErr || got != tt.want {
				t.Errorf("parseBrewVersion = %q, %v; want %q, error %t", got, err, tt.want, tt.wantErr)
			}
		})
	}
}
