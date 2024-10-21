package internal

import (
	"strings"
	"testing"
)

func Test_StatusModel(t *testing.T) {
	m := StatusModel{LastRound: 0}
	if !strings.Contains(m.String(), "LastRound: 0") {
		t.Fatal("expected \"LastRound: 0\", got ", m.String())
	}
}
