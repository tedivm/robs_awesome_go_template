package middleware

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Prometheus metrics.
var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

// uuidRegex matches UUID-formatted path segments for parameterization.
var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// MetricsHandler returns the Prometheus metrics HTTP handler.
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

// MetricsMiddleware returns a Chi-compatible middleware that records
// request counts and durations to Prometheus.
func MetricsMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap the response writer to capture the status code
			ww := &statusCaptureWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(ww, r)

			method := r.Method
			path := cleanPath(r.URL.Path)

			httpRequestsTotal.WithLabelValues(method, path, strconv.Itoa(ww.status)).Inc()
			httpRequestDuration.WithLabelValues(method, path).Observe(time.Since(start).Seconds())
		})
	}
}

// statusCaptureWriter wraps http.ResponseWriter to capture the status code.
type statusCaptureWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusCaptureWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// cleanPath removes query string, normalizes the path, and parameterizes
// UUIDs and numeric IDs to prevent unbounded Prometheus cardinality.
func cleanPath(p string) string {
	// Remove query string
	if i := strings.Index(p, "?"); i != -1 {
		p = p[:i]
	}
	// Parameterize UUIDs and numeric IDs in path segments
	parts := strings.Split(p, "/")
	for i, part := range parts {
		if part == "" {
			continue
		}
		if uuidRegex.MatchString(part) || isNumeric(part) {
			parts[i] = ":id"
		}
	}
	p = strings.Join(parts, "/")
	// Remove trailing slash (except root)
	if p != "/" && strings.HasSuffix(p, "/") {
		p = strings.TrimSuffix(p, "/")
	}
	return p
}

// isNumeric returns true if the string contains only digits.
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}
