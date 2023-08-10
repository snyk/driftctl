# Use bash syntax
SHELL=/bin/bash
# Go parameters
GOCMD=go
GOBINPATH=$(shell $(GOCMD) env GOPATH)/bin
GOMOD=$(GOCMD) mod
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=gotestsum
GOGET=$(GOCMD) get
GOINSTALL=$(GOCMD) install
GOTOOL=$(GOCMD) tool
GOFMT=$(GOCMD) fmt
# ACC tests params
ACC_PATTERN ?= TestAcc_

.PHONY: FORCE

.PHONY: all
all: fmt lint test build go.mod

.PHONY: build
build:
	SINGLE_TARGET=true ./scripts/build.sh

.PHONY: release
release:
	ENV=release ./scripts/build.sh

.PHONY: test
test:
	$(GOTEST) --format testname --junitfile unit-tests.xml -- -mod=readonly -coverprofile=cover.out.tmp -coverpkg=.,./pkg/... ./...
	cat cover.out.tmp | grep -v "mock_" > cover.out

.PHONY: coverage
coverage: test
	$(GOTOOL) cover -func=cover.out

.PHONY: acc
acc:
	DRIFTCTL_ACC=true $(GOTEST) --format standard-verbose --junitfile unit-tests-acc.xml -- -coverprofile=cover-acc.out -test.timeout 5h -coverpkg=./pkg/... -run=$(ACC_PATTERN) ./pkg/...

.PHONY: mocks
mocks:
	rm -rf mocks
	mockery --all


.PHONY: fmt
fmt:
	$(GOFMT) ./...

.PHONY: clean
clean:
	$(GOCLEAN)
	rm -f bin/*

.PHONY: lint
lint:
	@which golangci-lint > /dev/null 2>&1 || (curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | bash -s -- -b $(GOBINPATH) v1.46.2)
	golangci-lint run -v --timeout=10m

.PHONY: install-tools
install-tools:
	$(GOINSTALL) gotest.tools/gotestsum@v1.10.0
	$(GOINSTALL) github.com/vektra/mockery/v2@latest


go.mod: FORCE
	$(GOMOD) tidy
	$(GOMOD) verify
go.sum: go.mod
