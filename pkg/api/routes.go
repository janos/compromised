// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"fmt"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"resenje.org/web"
)

func newRouter(s *server) http.Handler {
	// Top level router
	baseRouter := http.NewServeMux()

	baseRouter.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "User-agent: *\nDisallow: /")
	})

	// Default route for all unknown paths
	baseRouter.HandleFunc("/", jsonNotFoundHandler)

	// API v1 router
	baseRouter.Handle("/v1/", newV1Router(s))

	// Final handler
	return web.ChainHandlers(
		handlers.CompressHandler,
		s.jsonRecoveryHandler,
		s.accessLogAndMetricsHandler,
		func(h http.Handler) http.Handler {
			return web.NewSetHeadersHandler(h, s.Headers)
		},
		web.FinalHandler(baseRouter),
	)
}

func newV1Router(s *server) http.Handler {
	r := mux.NewRouter().StrictSlash(true)
	r.UseEncodedPath()
	r.NotFoundHandler = http.HandlerFunc(jsonNotFoundHandler)

	r.Handle("/v1/passwords/{hash}", jsonMethodHandler{
		"GET": http.HandlerFunc(s.passwordHandler),
	})

	return web.ChainHandlers(
		jsonMaxBodyBytesHandler,
		web.NoCacheHeadersHandler,
		web.FinalHandler(r),
	)
}
