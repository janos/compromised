// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api_test

import (
	"context"
	"encoding/hex"
	"errors"
	"net/http"
	"testing"

	"resenje.org/compromised/pkg/api"
	mockpasswords "resenje.org/compromised/pkg/passwords/mock"
	"resenje.org/jsonhttp"
)

func TestPassword_notCompromised(t *testing.T) {
	sum := [20]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	var gotSum [20]byte
	c := newTestServer(t, testServerOptions{
		PasswordsService: mockpasswords.New(func(_ context.Context, s [20]byte) (uint64, error) {
			gotSum = s
			return 0, nil
		}),
	})

	var r api.PasswordResponse
	testResponseUnmarshal(t, c, http.MethodGet, "/v1/passwords/"+hex.EncodeToString(sum[:]), nil, http.StatusOK, &r)

	if gotSum != sum {
		t.Errorf("got sum %v, want %v", gotSum, sum)
	}

	if r.Compromised {
		t.Error("want not compromised")
	}

	if r.Count != 0 {
		t.Errorf("got count %v, want 0", r.Count)
	}
}

func TestPassword_compromised(t *testing.T) {
	sum := [20]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19}
	var gotSum [20]byte
	c := newTestServer(t, testServerOptions{
		PasswordsService: mockpasswords.New(func(_ context.Context, s [20]byte) (uint64, error) {
			gotSum = s
			return 10, nil
		}),
	})

	var r api.PasswordResponse
	testResponseUnmarshal(t, c, http.MethodGet, "/v1/passwords/"+hex.EncodeToString(sum[:]), nil, http.StatusOK, &r)

	if gotSum != sum {
		t.Errorf("got sum %v, want %v", gotSum, sum)
	}

	if !r.Compromised {
		t.Error("want compromised")
	}

	if r.Count != 10 {
		t.Errorf("got count %v, want 10", r.Count)
	}
}

func TestPassword_shortSum(t *testing.T) {
	c := newTestServer(t, testServerOptions{})

	testResponseDirect(t, c, http.MethodGet, "/v1/passwords/1234", nil, http.StatusNotFound, jsonhttp.StatusResponse{
		Code:    http.StatusNotFound,
		Message: http.StatusText(http.StatusNotFound),
	})
}

func TestPassword_invalidSum(t *testing.T) {
	c := newTestServer(t, testServerOptions{})

	testResponseDirect(t, c, http.MethodGet, "/v1/passwords/g1234567890abcdef1234567890abcdef1234567", nil, http.StatusNotFound, jsonhttp.StatusResponse{
		Code:    http.StatusNotFound,
		Message: http.StatusText(http.StatusNotFound),
	})
}

func TestPassword_error(t *testing.T) {
	c := newTestServer(t, testServerOptions{
		PasswordsService: mockpasswords.New(func(_ context.Context, s [20]byte) (uint64, error) {
			return 0, errors.New("test error")
		}),
	})

	testResponseDirect(t, c, http.MethodGet, "/v1/passwords/01234567890abcdef1234567890abcdef1234567", nil, http.StatusInternalServerError, jsonhttp.StatusResponse{
		Code:    http.StatusInternalServerError,
		Message: http.StatusText(http.StatusInternalServerError),
	})
}
