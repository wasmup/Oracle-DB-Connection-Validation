package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	OraclePingErrCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "oracle_db_connection_validation_ping_err",
	})

	OraclePingResponseTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "oracle_db_connection_validation_ping_second",
		Buckets: []float64{0.0001, 0.0002, 0.0005, 0.001, 0.002, 0.005, 0.010, 0.020, 0.050, 0.1, 0.2, 0.5, 1, 2, 5}, // 16 buckets
	})
)
