package internal

import (
	"context"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/algorandfoundation/hack-tui/api"
)

type MetricsModel struct {
	Enabled   bool
	Window    int
	RoundTime time.Duration
	TPS       float64
	RX        int
	TX        int
	LastTS    time.Time
	LastRX    int
	LastTX    int
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

// GetMetrics parses the /metrics endpoint from algod into a map
func GetMetrics(ctx context.Context, client *api.ClientWithResponses) (MetricsResponse, error) {
	res, err := client.MetricsWithResponse(ctx)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() != 200 {
		return nil, errors.New("invalid status code")
	}

	return parseMetricsContent(string(res.Body))
}
