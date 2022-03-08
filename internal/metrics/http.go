package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "eventdb",
		Name:      "request_duration",
		Help:      "duration per endpoint",
		Buckets:   []float64{0.01, 0.05, 0.1, 0.2, 0.4, 1, 2, 4, 8, 10, 20},
	}, []string{"endpoint"})
)
