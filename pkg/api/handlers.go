// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"net/http"
	"time"

	"resenje.org/jsonhttp"
	"resenje.org/web"
	accessLog "resenje.org/web/log/access"
	"resenje.org/web/recovery"
)

func (s *server) accessLogHandler(h http.Handler) http.Handler {
	return accessLog.NewHandler(h, s.AccessLogger)
}

func (s *server) jsonRecoveryHandler(h http.Handler) http.Handler {
	return recovery.New(h,
		recovery.WithLabel(s.Version),
		recovery.WithLogFunc(s.Logger.Errorf),
		recovery.WithPanicResponse(`{"message":"Internal Server Error","code":500}`, "application/json; charset=utf-8"),
	)
}

func (s *server) pageviewMetricsHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		s.metrics.PageviewCount.Inc()
		h.ServeHTTP(w, r)
		s.metrics.ResponseDuration.Observe(time.Since(start).Seconds())
	})
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
