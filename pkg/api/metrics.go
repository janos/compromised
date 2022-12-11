// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"github.com/prometheus/client_golang/prometheus"
	m "resenje.org/compromised/pkg/metrics"
)

type metrics struct {
	// declare metrics
	// must be exported to register them automatically using reflect
	PageviewCount    prometheus.Counter
	ResponseDuration prometheus.Histogram
	ResponseCount    *prometheus.CounterVec
}

func newMetrics() metrics {
	subsystem := "api"

	return metrics{
		PageviewCount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "request_count",
			Help:      "Number of API requests.",
		}),
		ResponseDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "response_duration_seconds",
			Help:      "Histogram of API response durations.",
			Buckets:   []float64{0.01, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		}),
		ResponseCount: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "response_code_count",
			Help:      "Number of responses by status codes from frontend router.",
		}, []string{"code"}),
	}
}
func (s *server) Metrics() (cs []prometheus.Collector) {
	return m.PrometheusCollectorsFromFields(s.metrics)
}
