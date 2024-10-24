package internal

import "time"

type MetricsModel struct {
	Window    int
	RoundTime time.Duration
	TPS       float64
	RX        int
	TX        int
}
