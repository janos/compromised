// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"resenje.org/compromised/pkg/passwords"
)

var _ passwords.Service = (*Service)(nil)

// Service implements passwords Service by communicating to the running
// 'compromised' API using HTTP client.
type Service struct {
	httpClient *http.Client
}

// New creates a new Service instance against the HTTP endpoint and with an
// optional custom HTTP client.
func New(endpoint string, httpClient *http.Client) (*Service, error) {
	baseURL, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	return &Service{
		httpClient: httpClientWithTransport(httpClient, baseURL),
	}, nil
}

type isPasswordCompromisedResponse struct {
	Compromised bool   `json:"compromised"`
	Count       uint64 `json:"count"`
}

// IsPasswordCompromised provides the information if the password is compromised
// by making an HTTP request to the running 'compromised' API.
func (s *Service) IsPasswordCompromised(ctx context.Context, sha1Sum [20]byte) (count uint64, err error) {
	var r isPasswordCompromisedResponse
	if err := s.request(ctx, http.MethodGet, "v1/passwords/"+hex.EncodeToString(sha1Sum[:]), &r); err != nil {
		return 0, err
	}

	if r.Compromised {
		return r.Count, nil
	}
	return 0, nil
}

func (s *Service) request(ctx context.Context, method, path string, v interface{}) error {
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)

	req.Header.Set("Accept", "application/json")

	r, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer drain(r.Body)

	if r.StatusCode != 200 {
		return fmt.Errorf("unexpected response status: %s", r.Status)
	}

	if v != nil && strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		return json.NewDecoder(r.Body).Decode(&v)
	}
	return nil
}

func httpClientWithTransport(c *http.Client, baseURL *url.URL) *http.Client {
	if c == nil {
		c = new(http.Client)
	}

	transport := c.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	if !strings.HasSuffix(baseURL.Path, "/") {
		baseURL.Path += "/"
	}

	c.Transport = roundTripperFunc(func(r *http.Request) (resp *http.Response, err error) {
		u, err := baseURL.Parse(r.URL.String())
		if err != nil {
			return nil, err
		}
		r.URL = u
		return transport.RoundTrip(r)
	})
	return c
}

// roundTripperFunc type is an adapter to allow the use of ordinary functions as
// http.RoundTripper interfaces. If f is a function with the appropriate
// signature, roundTripperFunc(f) is a http.RoundTripper that calls f.
type roundTripperFunc func(*http.Request) (*http.Response, error)

// RoundTrip calls f(r).
func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func drain(r io.ReadCloser) {
	go func() {
		// Panicking here does not put data in
		// an inconsistent state.
		defer func() {
			_ = recover()
		}()

		_, _ = io.Copy(io.Discard, r)
		_ = r.Close()
	}()
}
