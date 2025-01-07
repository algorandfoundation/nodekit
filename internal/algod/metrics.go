package algod

import (
	"context"
	"errors"
	"github.com/algorandfoundation/nodekit/api"
	"regexp"
	"strconv"
	"strings"
	"time"
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
	RX int

	// TX represents the number of bytes sent per second,
	// calculated from network metrics over a defined time interval.
	TX int

	// LastTS represents the timestamp of the last update to metrics,
	// used for calculating time deltas and rate metrics.
	LastTS time.Time

	// LastRX stores the total number of bytes received since the
	// last metrics update, used for RX rate calculation.
	LastRX int

	// LastTX stores the total number of bytes sent since the
	// last metrics update, used for TX rate calculation.
	LastTX int

	// Client provides an interface for interacting with API endpoints,
	// enabling metrics retrieval and other operations.
	Client api.ClientWithResponsesInterface

	// HttpPkg provides an interface for making HTTP requests,
	// facilitating communication with external APIs or services.
	HttpPkg api.HttpPkgInterface
}

// MetricsResponse represents a mapping of metric names to their integer values.
type MetricsResponse map[string]int

// parseMetricsContent parses Prometheus-style metrics content and returns a mapping of metric names to their integer values.
// It validates the input format, extracts key-value pairs, and handles errors during parsing.
func parseMetricsContent(content string) (MetricsResponse, error) {
	var err error
	result := MetricsResponse{}

	// Validate the Content
	var isValid bool
	isValid, err = regexp.MatchString(`^#`, content)
	isValid = isValid && err == nil && content != ""
	if !isValid {
		return nil, errors.New("invalid metrics content")
	}

	// Regex for Metrics Format,
	// selects all content that does not start with # in multiline mode
	re := regexp.MustCompile(`(?m)^[^#].*`)
	rows := re.FindAllString(content, -1)

	// Add the strings to the map
	for _, row := range rows {
		var value int
		keyValue := strings.Split(row, " ")
		value, err = strconv.Atoi(keyValue[1])
		result[keyValue[0]] = value
	}

	// Handle any error results
	if err != nil {
		return nil, err
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

	m.TX = max(0, int(float64(content["algod_network_sent_bytes_total"]-m.LastTX)/diff.Seconds()))
	m.RX = max(0, int(float64(content["algod_network_received_bytes_total"]-m.LastRX)/diff.Seconds()))

	m.LastTS = now
	m.LastTX = content["algod_network_sent_bytes_total"]
	m.LastRX = content["algod_network_received_bytes_total"]

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
		LastTS:    time.Time{},
		LastRX:    0,

		Client:  client,
		HttpPkg: httpPkg,
	}.Get(ctx, currentRound)
}
