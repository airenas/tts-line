package wrapservice

import (
	"github.com/prometheus/client_golang/prometheus"
)

var totalInvokeMetrics = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "avw_invoke_total",
		Help: "The total number of Model invocations",
	},
	[]string{"model", "voice"},
)

var totalRetryMetrics = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "avw_retry_total",
		Help: "The total number of Model invocation retries",
	},
	[]string{"model", "voice"},
)

var totalFailureMetrics = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "avw_failure_total",
		Help: "The total number of Model invocation failures",
	},
	[]string{"model", "voice"},
)

func init() {
	prometheus.MustRegister(totalInvokeMetrics, totalRetryMetrics, totalFailureMetrics)
}
