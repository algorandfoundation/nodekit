package linux

import (
	"reflect"
	"testing"
)

func TestAlgorandRepoAddCommand(t *testing.T) {
	tests := []struct {
		name string
		dnf5 bool
		want []string
	}{
		{
			name: "DNF 5",
			dnf5: true,
			want: []string{"sudo", "dnf5", "config-manager", "addrepo", "--from-repofile=" + algorandRPMRepoURL},
		},
		{
			name: "DNF 4",
			dnf5: false,
			want: []string{"sudo", "dnf", "config-manager", "--add-repo=" + algorandRPMRepoURL},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := algorandRepoAddCommand(tt.dnf5); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("algorandRepoAddCommand(%t) = %q, want %q", tt.dnf5, got, tt.want)
			}
		})
	}
}
