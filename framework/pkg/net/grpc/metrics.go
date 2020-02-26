package grpc

import "github.com/prometheus/client_golang/prometheus"

const (
	clientNamespace = "grpc_client"
	serverNamespace = "grpc_server"
)

var (
	_metricServerReqDur = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "grpc server requests duration(ms).",
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	}, []string{"method", "caller"})

	_metricServerReqCodeTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: serverNamespace,
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "grpc server requests code count.",
	}, []string{"method", "caller", "code"})

	_metricClientReqDur = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: clientNamespace,
		Subsystem: "requests",
		Name:      "duration_ms",
		Help:      "grpc client requests duration(ms).",
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000},
	}, []string{"method"})

	_metricClientReqCodeTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: clientNamespace,
		Subsystem: "requests",
		Name:      "code_total",
		Help:      "grpc client requests code count.",
	}, []string{"method", "code"})
)
