package algod

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/algorandfoundation/nodekit/api"
)

// Metrics represents runtime and performance metrics,
// including network traffic stats, TPS, and round time data.
type Metrics struct {
	// Enabled indicates whether metrics collection and processing are active.
	// If false, metrics are disabled or unavailable.
	Enabled bool

	// Window defines the range of rounds to consider when calculating metrics
	// such as TPS and average RoundTime.
	Window int

	// RoundTime represents the average duration of a round,
	// calculated based on recent round metrics.
	RoundTime time.Duration

	// TPS represents the calculated transactions per second,
	// based on the recent metrics over a defined window of rounds.
	TPS float64

	// RX represents the number of bytes received per second,
	// calculated from network metrics over a time interval.
	RX uint64

	// TX represents the number of bytes sent per second,
	// calculated from network metrics over a defined time interval.
	TX uint64

	// TXP2P represents the number of P2P bytes received per second,
	// calculated from network metrics over a time interval.
	RXP2P uint64

	// TXP2P represents the number of P2P bytes sent per second,
	// calculated from network metrics over a defined time interval.
	TXP2P uint64

	// LastTS represents the timestamp of the last update to metrics,
	// used for calculating time deltas and rate metrics.
	LastTS time.Time

	// LastRX stores the total number of bytes received since the
	// last metrics update, used for RX rate calculation.
	LastRX uint64

	// LastTX stores the total number of bytes sent since the
	// last metrics update, used for TX rate calculation.
	LastTX uint64

	// LastRXP2P stores the total number of P2P bytes received since
	// the last metrics update, used for RX rate calculation.
	LastRXP2P uint64

	// LastTXP2P stores the total number of P2P bytes sent since
	// the last metrics update, used for TX rate calculation.
	LastTXP2P uint64

	// Client provides an interface for interacting with API endpoints,
	// enabling metrics retrieval and other operations.
	Client api.ClientWithResponsesInterface

	// HttpPkg provides an interface for making HTTP requests,
	// facilitating communication with external APIs or services.
	HttpPkg api.HttpPkgInterface
}

// MetricsResponse represents a mapping of metric names to their integer values.
type MetricsResponse map[string]uint64

// parseMetricsContent parses Prometheus-style metrics content and returns a mapping of metric names to their integer values.
// It validates the input format, extracts key-value pairs, and handles errors during parsing.
func parseMetricsContent(content string) (MetricsResponse, error) {
	result := MetricsResponse{}

	// Validate the Content
	if !strings.HasPrefix(content, "#") || content == "" {
		return nil, errors.New("invalid metrics content: content must start with #")
	}

	// Regex for Metrics Format,
	// (.*?) - Capture metric key (name+labels, non-greedy)
	// \s    - Space delimiter
	// (.*?) - Capture metric value
	re := regexp.MustCompile(`(?m)^([^#].*?)\s(\d*?)$`)
	rows := re.FindAllStringSubmatch(content, -1)

	// Add the strings to the map
	for _, row := range rows {
		if len(row) < 3 {
			// Shouldn't happen given the regex above, but here as a sanity check
			continue
		}

		metricKey := strings.TrimSpace(row[1])
		valueStr := row[2]

		metricVal, err := strconv.ParseUint(valueStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse value '%s' for metric '%s': %w", valueStr, metricKey, err)
		}

		result[metricKey] = metricVal
	}

	// Give the user what they asked for
	return result, nil
}

// Get retrieves metrics data, processes network statistics,
// calculates TPS and round time, and updates the Metrics state.
func (m Metrics) Get(ctx context.Context, currentRound uint64) (Metrics, api.ResponseInterface, error) {
	response, err := m.Client.MetricsWithResponse(ctx)
	// Handle Errors and Status
	if err != nil {
		m.Enabled = false
		return m, response, err
	}
	if response.StatusCode() != 200 {
		m.Enabled = false
		return m, response, errors.New(InvalidStatus)
	}

	// Parse the Metrics Endpoint
	content, err := parseMetricsContent(string(response.Body))
	if err != nil {
		m.Enabled = false
		return m, response, err
	}

	// Handle Metrics
	m.Enabled = true
	now := time.Now()
	diff := now.Sub(m.LastTS)

	m.TX = max(0, uint64(float64(content["algod_network_sent_bytes_total"]-m.LastTX)/diff.Seconds()))
	m.RX = max(0, uint64(float64(content["algod_network_received_bytes_total"]-m.LastRX)/diff.Seconds()))

	m.TXP2P = max(0, uint64(float64(content["algod_network_p2p_sent_bytes_total"]-m.LastTXP2P)/diff.Seconds()))
	m.RXP2P = max(0, uint64(float64(content["algod_network_p2p_received_bytes_total"]-m.LastRXP2P)/diff.Seconds()))

	m.LastTS = now
	m.LastTX = content["algod_network_sent_bytes_total"]
	m.LastRX = content["algod_network_received_bytes_total"]
	m.LastTXP2P = content["algod_network_p2p_sent_bytes_total"]
	m.LastRXP2P = content["algod_network_p2p_received_bytes_total"]

	if int(currentRound) > m.Window {
		var blockMetrics BlockMetrics
		var blockMetricsResponse api.ResponseInterface
		blockMetrics, blockMetricsResponse, err = GetBlockMetrics(ctx, m.Client, currentRound, m.Window)
		if err != nil {
			return m, blockMetricsResponse, err
		}
		m.TPS = blockMetrics.TPS
		m.RoundTime = blockMetrics.AvgTime
	}

	return m, response, nil
}

// NewMetrics initializes and retrieves Metrics data by fetching the current round's metrics from the provided client.
// It requires a context, API client, HTTP package interface, and the current round number as inputs.
// Returns the populated Metrics instance, an API response interface, and an error, if any occurs.
func NewMetrics(ctx context.Context, client api.ClientWithResponsesInterface, httpPkg api.HttpPkgInterface, currentRound uint64) (Metrics, api.ResponseInterface, error) {
	return Metrics{
		Enabled:   false,
		Window:    100,
		RoundTime: 0 * time.Second,
		TPS:       0,
		RX:        0,
		TX:        0,
		RXP2P:     0,
		TXP2P:     0,
		LastTS:    time.Time{},
		LastTX:    0,
		LastRX:    0,
		LastTXP2P: 0,
		LastRXP2P: 0,

		Client:  client,
		HttpPkg: httpPkg,
	}.Get(ctx, currentRound)
}
