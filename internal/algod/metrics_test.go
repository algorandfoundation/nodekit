package algod

import (
	"context"
	"github.com/algorandfoundation/nodekit/api"
	"github.com/algorandfoundation/nodekit/internal/test"
	"strconv"
	"testing"
)

func Test_GetMetrics(t *testing.T) {
	client := test.GetClient(true)
	httpPkg := new(api.HttpPkg)
	metrics, _, err := NewMetrics(context.Background(), client, httpPkg, 0)
	if err == nil {
		t.Error("error expected")
	}

	metrics.Client = test.GetClient(false)
	metrics, _, err = metrics.Get(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}

	metrics.Client = test.NewClient(false, true)
	_, _, err = metrics.Get(context.Background(), 10)
	if err == nil {
		t.Error("expected error")
	}
}

func Test_parseMetrics(t *testing.T) {
	content := `# HELP algod_telemetry_drops_total telemetry messages dropped due to full queues
# TYPE algod_telemetry_drops_total counter
algod_telemetry_drops_total 0
# HELP algod_telemetry_errs_total telemetry messages dropped due to server error
# TYPE algod_telemetry_errs_total counter
algod_telemetry_errs_total 0
# HELP algod_ram_usage number of bytes runtime.ReadMemStats().HeapInuse
# TYPE algod_ram_usage gauge
algod_ram_usage 0
# HELP algod_crypto_vrf_generate_total Total number of calls to GenerateVRFSecrets
# TYPE algod_crypto_vrf_generate_total counter
algod_crypto_vrf_generate_total 0
# HELP algod_crypto_vrf_prove_total Total number of calls to VRFSecrets.Prove
# TYPE algod_crypto_vrf_prove_total counter
algod_crypto_vrf_prove_total 0
# HELP algod_crypto_vrf_hash_total Total number of calls to VRFProof.Hash
# TYPE algod_crypto_vrf_hash_total counter
algod_crypto_vrf_hash_total 0
`
	metrics, err := parseMetricsContent(content)

	if err != nil {
		t.Fatal(err)
	}

	if metrics["algod_telemetry_drops_total"] != 0 {
		t.Fatal(strconv.Itoa(metrics["algod_telemetry_drops_total"]) + " is not 0")
	}

	content = `INVALID`
	_, err = parseMetricsContent(content)
	if err == nil {
		t.Fatal(err)
	}

	content = `# HELP algod_telemetry_drops_total telemetry messages dropped due to full queues
# TYPE algod_telemetry_drops_total counter
algod_telemetry_drops_total NAN`
	_, err = parseMetricsContent(content)
	if err == nil {
		t.Fatal(err)
	}
}
