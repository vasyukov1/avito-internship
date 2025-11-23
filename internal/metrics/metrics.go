package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	HttpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	HttpRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	// Business metrics
	PRCreatedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "pr_created_total",
		Help: "Total number of created pull requests",
	}, []string{"team"})

	PRMergedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "pr_merged_total",
		Help: "Total number of merged pull requests",
	}, []string{"team"})

	PRReassignedTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "pr_reassigned_total",
		Help: "Total number of reviewer reassignments",
	}, []string{"team"})

	TeamsCreatedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "teams_created_total",
		Help: "Total number of created teams",
	})

	UsersActiveTotal = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "users_active_total",
		Help: "Total number of active users",
	}, []string{"team"})

	// Error metrics
	ErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "errors_total",
		Help: "Total number of errors by type",
	}, []string{"type", "operation"})

	// Database metrics
	DBQueriesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "db_queries_total",
		Help: "Total number of database queries",
	}, []string{"operation"})

	DBQueryDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "db_query_duration_seconds",
		Help:    "Database query duration in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation"})
)
