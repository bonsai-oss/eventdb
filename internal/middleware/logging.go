package middleware

import (
	"github.com/prometheus/client_golang/prometheus"
	"golang.fsrv.services/eventdb/internal/metrics"
	"log"
	"net/http"
	"time"
)

func Logging(logger *log.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// serve http request
			next.ServeHTTP(w, r)
			duration := time.Since(start)
			metrics.RequestDuration.With(prometheus.Labels{"endpoint": r.RequestURI}).Observe(duration.Seconds())

			logger.Printf("%v %v %v", r.Method, r.RequestURI, duration)
		})
	}
}
