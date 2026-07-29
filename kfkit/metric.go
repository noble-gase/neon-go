package kfkit

import "github.com/prometheus/client_golang/prometheus"

const namespace = "kafka_client"

var (
	metricReqDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "requests",
			Name:      "duration_ms",
			Help:      "kafka pub/sub requests duration(ms).",
			Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000, 2500},
		}, []string{"topic", "command"},
	)

	metricResult = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "process_result",
			Help:      "kafka pub/sub result",
		}, []string{"topic", "command", "result"},
	)

	metricSubDelay = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "delay",
			Name:      "duration_ms",
			Help:      "kafka sub delay time(ms).",
			Buckets:   []float64{10, 50, 200, 1000, 5000, 20000, 100000},
		}, []string{"topic"},
	)
)

func init() {
	prometheus.MustRegister(metricReqDuration)
	prometheus.MustRegister(metricResult)
	prometheus.MustRegister(metricSubDelay)
}
