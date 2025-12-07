package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

// Метрики
var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_time_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: []float64{0.1, 0.3, 0.5, 1, 2, 5},
		},
		[]string{"method", "path"},
	)

	activeRequests = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_active_requests",
			Help: "Current number of active requests",
		},
	)

	errorsCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_errors_total",
			Help: "Total application errors",
		},
		[]string{"type", "component"},
	)
)

// Middleware для сбора метрик
func (m *MiddlewareManager) MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		activeRequests.Inc()
		defer activeRequests.Dec()

		rw := &responseWriter{w, http.StatusOK}
		defer func() {
			duration := time.Since(start).Seconds()

			httpRequestsTotal.WithLabelValues(
				r.Method,
				r.URL.Path,
				http.StatusText(rw.status),
			).Inc()

			if rw.status >= 400 {
				errorsCounter.WithLabelValues(strconv.Itoa(rw.status)).Inc()
			}

			httpDuration.WithLabelValues(
				r.Method,
				r.URL.Path,
			).Observe(duration)
		}()

		next.ServeHTTP(w, r)
	})
}
