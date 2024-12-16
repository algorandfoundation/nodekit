package api

import (
	"testing"
)

func Test_GetLatestCatchpoint(t *testing.T) {
	catchpoint, err := GetLatestCatchpointWithResponse(new(HttpPkg), "mainnet")
	if err != nil {
		t.Error(err)
	}
	t.Log(catchpoint)
}
