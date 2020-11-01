# Copyright (c) 2020, Compromised AUTHORS.
# All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

GO ?= go
GOLANGCI_LINT ?= $$($(GO) env GOPATH)/bin/golangci-lint
GOLANGCI_LINT_VERSION ?= v1.32.1

COMMIT ?= "$(shell git describe --long --dirty --always --match "" || true)"
LDFLAGS ?= -s -w -X resenje.org/compromised.commit="$(COMMIT)"

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
lint: linter
	$(GOLANGCI_LINT) run

.PHONY: linter
linter:
	test -f $(GOLANGCI_LINT) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$($(GO) env GOPATH)/bin $(GOLANGCI_LINT_VERSION)

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
