package http

import (
	"github.com/prometheus/client_golang/prometheus"
)

var totalRequests = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "api_http_request_count",
		Help: "Number of http requests",
	},
	[]string{"path"},
)

var recoveredPanics = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "api_recovered_panics_count",
		Help: "Number of recovered panics",
	},
	[]string{"error"},
)

func initMetrics() error {
	if err := prometheus.Register(totalRequests); err != nil {
		return err
	}

	if err := prometheus.Register(recoveredPanics); err != nil {
		return err
	}

	return nil
}
