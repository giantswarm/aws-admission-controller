package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	metricNamespace = "aws_admission_controller"
	metricSubsystem = "webhook"
)

var (
	labels = []string{"webhook", "resource"}

	DurationRequests = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: metricNamespace,
		Subsystem: metricSubsystem,
		Name:      "request_duration_seconds",
		Help:      "Duration of request",
		Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, .75, 1, 1.25, 1.5, 2, 2.5, 5, 10},
	}, labels)
	InternalError = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricNamespace,
		Subsystem: metricSubsystem,
		Name:      "errors_total",
		Help:      "Total number of errors",
	}, labels)
	InvalidRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricNamespace,
		Subsystem: metricSubsystem,
		Name:      "requests_invalid_total",
		Help:      "Total number of invalid requests",
	}, labels)
	RejectedRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricNamespace,
		Subsystem: metricSubsystem,
		Name:      "requests_rejected_total",
		Help:      "Total number of rejected requests",
	}, labels)
	SuccessfulRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricNamespace,
		Subsystem: metricSubsystem,
		Name:      "requests_successful_total",
		Help:      "Total number of successful requests",
	}, labels)
	TotalRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: metricNamespace,
		Subsystem: metricSubsystem,
		Name:      "requests_total",
		Help:      "Total number of requests",
	}, labels)
)

func init() {
	prometheus.MustRegister(TotalRequests, InvalidRequests, RejectedRequests, SuccessfulRequests, DurationRequests)
}
