// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package http_test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	httppasswords "resenje.org/compromised/pkg/passwords/http"
)

func TestIsPasswordCompromised_compromised(t *testing.T) {
	client, mux := newClient(t)

	hash := "3d5896ffe806a482490b99f690650995b63c3513"
	var want uint64 = 101

	mux.HandleFunc("/v1/passwords/"+hash, func(w http.ResponseWriter, r *http.Request) {
		b, err := json.Marshal(httppasswords.IsPasswordCompromisedResponse{
			Compromised: true,
			Count:       want,
		})
		if err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", jsonContentType)
		_, _ = w.Write(b)
	})

	got, err := client.IsPasswordCompromised(context.Background(), hexDecodeSHA1Sum(t, hash))
	if err != nil {
		t.Fatal(err)
	}

	if got != want {
		t.Errorf("got count %v, want %v", got, want)
	}
}

func TestIsPasswordCompromised_notCompromised(t *testing.T) {
	client, mux := newClient(t)

	hash := "3d5896ffe806a482490b99f690650995b63c3514"
	var want uint64

	mux.HandleFunc("/v1/passwords/"+hash, func(w http.ResponseWriter, r *http.Request) {
		b, err := json.Marshal(httppasswords.IsPasswordCompromisedResponse{
			Compromised: false,
			Count:       want,
		})
		if err != nil {
			t.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", jsonContentType)
		_, _ = w.Write(b)
	})

	got, err := client.IsPasswordCompromised(context.Background(), hexDecodeSHA1Sum(t, hash))
	if err != nil {
		t.Fatal(err)
	}

	if got != want {
		t.Errorf("got count %v, want %v", got, want)
	}
}

const jsonContentType = "application/json; charset=utf-8"

func newClient(t testing.TB) (client *httppasswords.Service, mux *http.ServeMux) {
	t.Helper()

	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client, err := httppasswords.New(server.URL, server.Client())
	if err != nil {
		t.Fatal(err)
	}

	return client, mux
}

func hexDecodeSHA1Sum(t *testing.T, s string) (sum [20]byte) {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	copy(sum[:], b)
	return sum
}
