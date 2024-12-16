package internal

import (
	"context"
	"testing"
)

func Test_GetLatestCatchpoint(t *testing.T) {
	catchpoint, err := GetLatestCatchpoint(context.Background(), new(HttpPkg), "mainnet")
	if err != nil {
		t.Error(err)
	}
	t.Log(catchpoint)
}
