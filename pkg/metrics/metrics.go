// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package metrics

import (
	"reflect"

	"github.com/prometheus/client_golang/prometheus"
)

var Namespace = "compromised"

type Collector interface {
	Metrics() []prometheus.Collector
}

func PrometheusCollectorsFromFields(i interface{}) (cs []prometheus.Collector) {
	v := reflect.Indirect(reflect.ValueOf(i))
	for i := 0; i < v.NumField(); i++ {
		if !v.Field(i).CanInterface() {
			continue
		}
		if u, ok := v.Field(i).Interface().(prometheus.Collector); ok {
			cs = append(cs, u)
		}
	}
	return cs
}
