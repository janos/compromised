// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"net/http"
	"strings"

	"github.com/felixge/httpsnoop"
	"resenje.org/jsonhttp"
	"resenje.org/logging"
	"resenje.org/web"
	"resenje.org/web/recovery"
)

func (s *server) accessLogAndMetricsHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.metrics.PageviewCount.Inc()
		m := httpsnoop.CaptureMetrics(h, w, r)
		referrer := r.Referer()
		if referrer == "" {
			referrer = "-"
		}
		userAgent := r.UserAgent()
		if userAgent == "" {
			userAgent = "-"
		}
		ips := []string{}
		if v := r.Header.Get(s.RealIPHeaderName); v != "" {
			ips = append(ips, v)
		}
		xips := "-"
		if len(ips) > 0 {
			xips = strings.Join(ips, ", ")
		}
		status := m.Code
		var level logging.Level
		switch {
		case status >= 500:
			level = logging.ERROR
		case status >= 400:
			level = logging.WARNING
		case status >= 300:
			level = logging.INFO
		case status >= 200:
			level = logging.INFO
		default:
			level = logging.DEBUG
		}
		s.AccessLogger.Logf(level, "%s \"%s\" \"%v %s %v\" %d %d %f \"%s\" \"%s\"", r.RemoteAddr, xips, r.Method, r.RequestURI, r.Proto, status, m.Written, m.Duration.Seconds(), referrer, userAgent)

		s.metrics.ResponseDuration.Observe(m.Duration.Seconds())
	})
}

func (s *server) jsonRecoveryHandler(h http.Handler) http.Handler {
	return recovery.New(h,
		recovery.WithLabel(s.Version),
		recovery.WithLogFunc(s.Logger.Errorf),
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
