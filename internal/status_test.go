package internal

import (
	"testing"
)

func Test_StatusModel(t *testing.T) {
	m := StatusModel{LastRound: 0}
	if m.String() != "Last round: 0" {
		t.Fatal("expected \"Last round: 0\", got ", m.String())
	}
}
