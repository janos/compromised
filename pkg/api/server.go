// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/exp/slog"
	"resenje.org/compromised/pkg/passwords"
	"resenje.org/recovery"
)

// Handler is an http.Handler with exposed metrics.
type Handler interface {
	http.Handler
	Metrics() (cs []prometheus.Collector)
}

type server struct {
	Options

	handler http.Handler
	metrics metrics
}

// Options structure contains optional properties for the Handler.
type Options struct {
	Version          string
	Headers          map[string]string
	RealIPHeaderName string

	Logger       *slog.Logger
	AccessLogger *slog.Logger

	RecoveryService *recovery.Service

	PasswordsService passwords.Service
}

// New initializes a new Handler with provided options.
func New(o Options) (h Handler, err error) {
	if o.Version == "" {
		o.Version = "0"
	}
	s := &server{
		Options: o,
		metrics: newMetrics(),
	}

	s.handler = newRouter(s)

	return s, nil
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.handler.ServeHTTP(w, r)
}
