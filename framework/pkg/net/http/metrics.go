package khttp

import "github.com/prometheus/client_golang/prometheus"

const (
	clientNamespace = "http_client"
	serverNamespace = "http_server"
)

var (
	_metricServerReqDur = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "http server requests duration(ms).",
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	}, []string{"path", "caller", "method"})

	_metricServerReqCodeTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "http server requests code count.",
	}, []string{"path", "caller", "method", "code"})

	_metricServerBBR = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: clientNamespace,
		Subsystem: "requests",
		Name:      "bbr_total",
		Help:      "http client requests code count.",
	}, []string{"url", "method"})

	_metricClientReqDur = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: clientNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "http client requests duration(ms).",
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	}, []string{"path", "method"})

	_metricClientReqCodeTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: clientNamespace,
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "http client requests code count.",
	}, []string{"path", "method", "code"})
)
