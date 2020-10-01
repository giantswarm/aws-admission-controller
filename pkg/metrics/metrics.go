package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	MetricNamespace = "aws_admission_controller"
	metricSubsystem = "webhook"
)

var (
	labels = []string{"webhook", "resource"}

	InternalError = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: MetricNamespace,
		Subsystem: metricSubsystem,
		Name:      "internal_error",
		Help:      "Total number of errors",
	}, labels)
	ApprovedRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: MetricNamespace,
		Subsystem: metricSubsystem,
		Name:      "successful_requests",
		Help:      "Total number of successful requests",
	}, labels)
	DurationRequests = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: MetricNamespace,
		Subsystem: metricSubsystem,
		Name:      "duration_requests",
		Help:      "Duration of requests",
		Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, .75, 1, 1.25, 1.5, 2, 2.5, 5, 10},
	}, labels)
	InvalidRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: MetricNamespace,
		Subsystem: metricSubsystem,
		Name:      "invalid_requests",
		Help:      "Total number of invalid requests",
	}, labels)
	RejectedRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: MetricNamespace,
		Subsystem: metricSubsystem,
		Name:      "rejected_requests",
		Help:      "Total number of rejected requests",
	}, labels)
	TotalRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: MetricNamespace,
		Subsystem: metricSubsystem,
		Name:      "requests",
		Help:      "Total number of requests",
	}, labels)
)

func init() {
	prometheus.MustRegister(TotalRequests, InvalidRequests, RejectedRequests, ApprovedRequests, DurationRequests)
}
