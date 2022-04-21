# Copyright (c) 2020, Compromised AUTHORS.
# All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

GO ?= go
GOLANGCI_LINT ?= $$($(GO) env GOPATH)/bin/golangci-lint

LDFLAGS ?= -s -w
TAGS += timetzdata

.PHONY: all
all: binary lint vet test

.PHONY: develop
develop: binary run

.PHONY: binary
binary: export CGO_ENABLED=0
binary: dist FORCE
	$(GO) version
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o dist/compromised ./cmd/compromised

dist:
	mkdir $@

.PHONY: lint
lint:
	$(GOLANGCI_LINT) run

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: test
test:
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" ./...
	$(GO) test -race -v ./...

.PHONY: run
run:
	./dist/compromised config
	./dist/compromised

.PHONY: clean
clean:
	$(GO) clean
	rm -rf dist/

FORCE:
