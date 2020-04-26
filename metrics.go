package zergrepo

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	methodLabel = "method"
)

// Metric is responsible for collecting metrics and registering them.
type Metric struct {
	callTotal    *prometheus.CounterVec
	callErrTotal *prometheus.CounterVec
	callDuration *prometheus.HistogramVec
}

// MustMetric collects a new metric instance and automatically registers the metrics.
// If it fails to register a metric, a panic will occur.
func MustMetric(namespace, subsystem string) *Metric {
	m, err := NewMetric(namespace, subsystem, prometheus.DefaultRegisterer)
	if err != nil {
		panic(err)
	}
	return m
}

// NewMetric is gathering a new metric instantiation. Returns an error if a metric failed to be registered.
func NewMetric(namespace, subsystem string, reg prometheus.Registerer) (*Metric, error) {
	m := &Metric{}

	m.callTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "call_total",
			Help:      "Amount of repo calls.",
		},
		[]string{methodLabel},
	)

	m.callErrTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "errors_total",
			Help:      "Amount of repo errors.",
		},
		[]string{methodLabel},
	)

	m.callDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "call_duration_seconds",
			Help:      "Repo call latency.",
		},
		[]string{methodLabel},
	)

	collectors := []prometheus.Collector{m.callTotal, m.callErrTotal, m.callDuration}
	for i := range collectors {
		err := reg.Register(collectors[i])
		if err != nil {
			return nil, fmt.Errorf("register metric: %w", err)
		}
	}

	return m, nil
}

func (m *Metric) collect(method string, fn func() error) func() error {
	return func() (err error) {
		start := time.Now()
		l := prometheus.Labels{methodLabel: method}

		defer func() {
			m.callTotal.With(l).Inc()
			m.callDuration.With(l).Observe(time.Since(start).Seconds())

			if err != nil {
				m.callErrTotal.With(l).Inc()
			} else if err := recover(); err != nil {
				m.callErrTotal.With(l).Inc()
				panic(err)
			}
		}()

		return fn()
	}
}
