// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package file

import (
	"github.com/prometheus/client_golang/prometheus"
	m "resenje.org/compromised/pkg/metrics"
)

type metrics struct {
	// all metrics fields must be exported
	// to be able to return them by Metrics()
	// using reflection
	CheckedCount     prometheus.Counter
	CompromisedCount prometheus.Counter
}

func newMetrics() metrics {
	subsystem := "passwords"

	return metrics{
		CheckedCount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "checked_count",
			Help:      "Number of checked passwords.",
		}),
		CompromisedCount: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: m.Namespace,
			Subsystem: subsystem,
			Name:      "compromised_count",
			Help:      "Number of detected compromised passwords.",
		}),
	}
}

// Metrics provides prometheus metrics from this Service.
func (s *Service) Metrics() (cs []prometheus.Collector) {
	return m.PrometheusCollectorsFromFields(s.metrics)
}
