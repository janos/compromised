// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"encoding/hex"
	"net/http"

	"github.com/gorilla/mux"
	"resenje.org/jsonhttp"
)

type passwordResponse struct {
	Compromised bool   `json:"compromised"`
	Count       uint64 `json:"count,omitempty"`
}

func (s *server) passwordHandler(w http.ResponseWriter, r *http.Request) {
	hash := mux.Vars(r)["hash"]

	if len(hash) != 40 {
		jsonhttp.NotFound(w, nil)
		return
	}

	slice, err := hex.DecodeString(hash)
	if err != nil {
		jsonhttp.NotFound(w, nil)
		return
	}

	var sum [20]byte
	copy(sum[:], slice)

	count, err := s.PasswordsService.IsPasswordCompromised(r.Context(), sum)
	if err != nil {
		s.Logger.Error("api password handler: is password compromised", err, "hash", hash)
		jsonhttp.InternalServerError(w, nil)
		return
	}

	jsonhttp.OK(w, passwordResponse{
		Compromised: count > 0,
		Count:       count,
	})
}
