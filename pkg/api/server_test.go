// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"golang.org/x/exp/slog"
	"resenje.org/compromised"
	"resenje.org/compromised/pkg/api"
	"resenje.org/compromised/pkg/passwords"
	"resenje.org/recovery"
	"resenje.org/web"
)

type testServerOptions struct {
	PasswordsService passwords.Service
}

func newTestServer(t *testing.T, o testServerOptions) *http.Client {
	version := "0.1.0-test"
	logger := slog.Default()

	s, err := api.New(api.Options{
		Version:      version,
		Logger:       logger,
		AccessLogger: logger,
		RecoveryService: &recovery.Service{
			Version: compromised.Version(),
		},
		PasswordsService: o.PasswordsService,
	})
	if err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(s)
	t.Cleanup(ts.Close)

	httpClient := &http.Client{
		Transport: web.RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
			u, err := url.Parse(ts.URL + r.URL.String())
			if err != nil {
				return nil, err
			}
			r.URL = u
			return ts.Client().Transport.RoundTrip(r)
		}),
	}

	return httpClient
}

func testResponseUnmarshal(t *testing.T, client *http.Client, method, url string, body io.Reader, responseCode int, response interface{}) {
	t.Helper()

	resp, err := request(client, method, url, body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != responseCode {
		t.Fatalf("got response status %s, want %v %s", resp.Status, responseCode, http.StatusText(responseCode))
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
}

func testResponseDirect(t *testing.T, client *http.Client, method, url string, body io.Reader, responseCode int, response interface{}) {
	t.Helper()

	resp, err := request(client, method, url, body)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != responseCode {
		t.Fatalf("got response status %s, want %v %s", resp.Status, responseCode, http.StatusText(responseCode))
	}

	gotBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	wantBytes, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}

	got := string(bytes.TrimSpace(gotBytes))
	want := string(wantBytes)

	if got != want {
		t.Fatalf("got response %s, want %s", got, want)
	}
}

func request(client *http.Client, method, url string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	return client.Do(req)
}
