package types

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	functionNameLabel = "function_name"
)

type MetricsCollector interface {
	RegisterFunctionInvocation(fn string)
	RegisterTopicMapSync()
	Serve(bindAddr string) error
}

type defaultMetricsCollector struct {
	lastFunctionInvocationTimestamp *prometheus.GaugeVec
	lastTopicMapSyncTimestamp       prometheus.Gauge
	registry                        *prometheus.Registry
	totalFunctionInvocations        *prometheus.CounterVec
	totalTopicMapSyncs              prometheus.Counter
}

func newDefaultMetricsCollector() *defaultMetricsCollector {
	cm := &defaultMetricsCollector{
		lastFunctionInvocationTimestamp: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Help:      "The timestamp at which a function has last been invoked.",
			Name:      "last_function_invocation_timestamp",
			Namespace: "controller",
		}, []string{functionNameLabel}),
		lastTopicMapSyncTimestamp: prometheus.NewGauge(prometheus.GaugeOpts{
			Help:      "The timestamp at which the topic map has last been synced.",
			Name:      "last_topic_map_sync_timestamp",
			Namespace: "controller",
		}),
		registry: prometheus.NewRegistry(),
		totalFunctionInvocations: prometheus.NewCounterVec(prometheus.CounterOpts{
			Help:      "The number of times a function has last been invoked.",
			Name:      "total_function_invocations",
			Namespace: "controller",
		}, []string{functionNameLabel}),
		totalTopicMapSyncs: prometheus.NewCounter(prometheus.CounterOpts{
			Help:      "The number of times the topic map has been synced.",
			Name:      "total_topic_map_syncs",
			Namespace: "controller",
		}),
	}
	cm.registry.MustRegister(cm.lastFunctionInvocationTimestamp)
	cm.registry.MustRegister(cm.lastTopicMapSyncTimestamp)
	cm.registry.MustRegister(cm.totalFunctionInvocations)
	cm.registry.MustRegister(cm.totalTopicMapSyncs)
	return cm
}

func (cm *defaultMetricsCollector) RegisterFunctionInvocation(fn string) {
	cm.lastFunctionInvocationTimestamp.WithLabelValues(fn).SetToCurrentTime()
	cm.totalFunctionInvocations.WithLabelValues(fn).Inc()
}

func (cm *defaultMetricsCollector) RegisterTopicMapSync() {
	cm.lastTopicMapSyncTimestamp.SetToCurrentTime()
	cm.totalTopicMapSyncs.Inc()
}

func (cm *defaultMetricsCollector) Serve(bindAddr string) error {
	m := http.NewServeMux()
	m.Handle("/metrics", promhttp.HandlerFor(cm.registry, promhttp.HandlerOpts{}))
	s := http.Server{Addr: bindAddr, Handler: m}
	return s.ListenAndServe()
}
