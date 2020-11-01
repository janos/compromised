// Copyright (c) 2020, Compromised AUTHORS.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mock

import (
	"context"

	"resenje.org/compromised/pkg/passwords"
)

var _ passwords.Service = (*Service)(nil)

// Service implements passwords service with injectable functionality mainly
// meant unit testing services that depend on passwords service.
type Service struct {
	isPasswordCompromisedFunc func(ctx context.Context, sha1Sum [20]byte) (uint64, error)
}

// New creates a new instance of Service by injecting the passed function as the
// service method.
func New(isPasswordCompromisedFunc func(ctx context.Context, sha1Sum [20]byte) (uint64, error)) *Service {
	return &Service{
		isPasswordCompromisedFunc: isPasswordCompromisedFunc,
	}
}

// IsPasswordCompromised calls the function what is passed to the New
// constructor.
func (s *Service) IsPasswordCompromised(ctx context.Context, sha1Sum [20]byte) (uint64, error) {
	return s.isPasswordCompromisedFunc(ctx, sha1Sum)
}
