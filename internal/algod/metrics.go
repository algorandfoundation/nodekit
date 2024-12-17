package algod

import (
	"context"
	"errors"
	"github.com/algorandfoundation/algorun-tui/api"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Metrics struct {
	Enabled   bool
	Window    int
	RoundTime time.Duration
	TPS       float64
	RX        int
	TX        int
	LastTS    time.Time
	LastRX    int
	LastTX    int

	Client  api.ClientWithResponsesInterface
	HttpPkg api.HttpPkgInterface
}

type MetricsResponse map[string]int

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

func NewMetrics(ctx context.Context, client api.ClientWithResponsesInterface, httpPkg api.HttpPkgInterface, currentRound uint64) (Metrics, api.ResponseInterface, error) {
	return Metrics{
		Enabled:   false,
		Window:    100,
		RoundTime: 0 * time.Second,
		TPS:       0,
		RX:        0,
		TX:        0,
		LastTS:    time.Now().Add(-1_000_000 * time.Hour),
		LastRX:    0,

		Client:  client,
		HttpPkg: httpPkg,
	}.Get(ctx, currentRound)
}
