package metrics

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Names of metrics
const (
	RequestsTotal      = "wave_http_requests_total"
	RequestDuration    = "wave_http_request_duration_seconds"
	ResponseSize       = "wave_http_response_size_bytes"
	DatabaseOperations = "wave_database_operations_total"
	DatabaseDuration   = "wave_database_operation_duration_seconds"
	ActiveConnections  = "wave_active_connections"
	MessageCount       = "wave_messages_total"
	ErrorsTotal        = "wave_errors_total"
)

var (
	// Registry for all metrics
	registry = prometheus.NewRegistry()

	// HTTP metrics
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: RequestsTotal,
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    RequestDuration,
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	responseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    ResponseSize,
			Help:    "HTTP response size in bytes",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000},
		},
		[]string{"method", "path"},
	)

	// Database metrics
	databaseOperations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: DatabaseOperations,
			Help: "Total number of database operations",
		},
		[]string{"operation", "table"},
	)

	databaseDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    DatabaseDuration,
			Help:    "Database operation duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	// Application metrics
	activeConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: ActiveConnections,
			Help: "Current number of active connections",
		},
	)

	messageCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: MessageCount,
			Help: "Total number of messages processed",
		},
		[]string{"type"},
	)

	errorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: ErrorsTotal,
			Help: "Total number of errors",
		},
		[]string{"type"},
	)
)

func init() {
	// Register metrics with the registry
	registry.MustRegister(requestsTotal)
	registry.MustRegister(requestDuration)
	registry.MustRegister(responseSize)
	registry.MustRegister(databaseOperations)
	registry.MustRegister(databaseDuration)
	registry.MustRegister(activeConnections)
	registry.MustRegister(messageCount)
	registry.MustRegister(errorsTotal)
}

// RegisterMetricsHandler registers the metrics endpoint with Echo
func RegisterMetricsHandler(e *echo.Echo) {
	e.GET("/metrics", echo.WrapHandler(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))
}

// RecordRequestMetrics records metrics for an HTTP request
func RecordRequestMetrics(method, path string, status int, duration float64, size int) {
	requestsTotal.WithLabelValues(method, path, fmt.Sprintf("%d", status)).Inc()
	requestDuration.WithLabelValues(method, path).Observe(duration)
	responseSize.WithLabelValues(method, path).Observe(float64(size))
}

// RecordDatabaseMetrics records metrics for a database operation
func RecordDatabaseMetrics(operation, table string, duration float64) {
	databaseOperations.WithLabelValues(operation, table).Inc()
	databaseDuration.WithLabelValues(operation, table).Observe(duration)
}

// RecordActiveConnection records an active connection
func RecordActiveConnection(delta int) {
	activeConnections.Add(float64(delta))
}

// RecordMessage records a message
func RecordMessage(messageType string) {
	messageCount.WithLabelValues(messageType).Inc()
}

// RecordError records an error
func RecordError(errorType string) {
	errorsTotal.WithLabelValues(errorType).Inc()
}
