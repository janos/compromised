// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api_test

import (
	"io"
	"net/http"
	"testing"

	"resenje.org/jsonhttp"
)

func TestRoot(t *testing.T) {
	c := newTestServer(t, testServerOptions{})

	testResponseDirect(t, c, http.MethodGet, "/", nil, http.StatusNotFound, jsonhttp.StatusResponse{
		Code:    http.StatusNotFound,
		Message: http.StatusText(http.StatusNotFound),
	})
}

func TestRobotsTxt(t *testing.T) {
	c := newTestServer(t, testServerOptions{})

	resp, err := request(c, http.MethodGet, "/robots.txt", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	got := string(b)
	want := "User-agent: *\nDisallow: /\n"

	if got != want {
		t.Errorf("got response %q, want %q", got, want)
	}
}
