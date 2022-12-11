// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"net/http"
	"strconv"
	"time"

	"resenje.org/jsonhttp"
	"resenje.org/web"
	"resenje.org/web/logging"
	"resenje.org/web/recovery"
)

func (s *server) accessLogAndMetricsHandler(h http.Handler) http.Handler {
	return logging.NewAccessLogHandler(h, s.AccessLogger, &logging.AccessLogOptions{
		RealIPHeaderName: s.RealIPHeaderName,
		PreHook: func(_ http.ResponseWriter, _ *http.Request) {
			s.metrics.PageviewCount.Inc()
		},
		PostHook: func(code int, duration time.Duration, _ int64) {
			s.metrics.ResponseDuration.Observe(duration.Seconds())
			s.metrics.ResponseCount.WithLabelValues(strconv.Itoa(code)).Inc()
		},
		LogMessage: "api access",
	})
}

func (s *server) jsonRecoveryHandler(h http.Handler) http.Handler {
	return recovery.New(h,
		recovery.WithLabel(s.Version),
		recovery.WithLogger(s.Logger),
		recovery.WithPanicResponse(`{"message":"Internal Server Error","code":500}`, "application/json; charset=utf-8"),
	)
}

func jsonMaxBodyBytesHandler(h http.Handler) http.Handler {
	return web.MaxBodyBytesHandler{
		Handler: h,
		Limit:   2 * 1024 * 1024,
		BodyFunc: func(r *http.Request) (string, error) {
			return `{"message":"Request Entity Too Large","code":413}`, nil
		},
		ContentType:  "application/json; charset=utf-8",
		ErrorHandler: nil,
	}
}

func jsonNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	jsonhttp.NotFound(w, nil)
}

type jsonMethodHandler map[string]http.Handler

func (h jsonMethodHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	web.HandleMethods(h, `{"message":"Method Not Allowed","code":405}`, "application/json; charset=utf-8", w, r)
}
