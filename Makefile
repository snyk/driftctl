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
GOTOOL=$(GOCMD) tool
GOFMT=$(GOCMD) fmt
# ACC tests params
ACC_PATTERN ?= TestAcc_

.PHONY: FORCE

.PHONY: all
all: fmt lint test build

.PHONY: build
build:
	scripts/build.sh

.PHONY: release
release:
	ENV=release scripts/build.sh

.PHONY: test
test: deps
	$(GOTEST) --format testname --junitfile unit-tests.xml -- -coverprofile=cover.out -coverpkg=.,./pkg/... ./...

.PHONY: coverage
coverage: test
	$(GOTOOL) cover -func=cover.out

.PHONY: acc
acc:
	DRIFTCTL_ACC=true $(GOTEST) --format testname --junitfile unit-tests-acc.xml -- -coverprofile=cover-acc.out -test.timeout 1h -coverpkg=./pkg/... -run=$(ACC_PATTERN) ./pkg/...

.PHONY: mocks
mocks: deps
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
	@which golangci-lint > /dev/null 2>&1 || (curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | bash -s -- -b $(GOBINPATH) v1.31.0)
	golangci-lint run -v --timeout=4m

.PHONY: deps
deps:
	$(GOMOD) download

.PHONY: install-tools
install-tools:
	$(GOGET) -u gotest.tools/gotestsum
	$(GOGET) github.com/vektra/mockery/.../


go.mod: FORCE
	$(GOMOD) tidy
	$(GOMOD) verify
go.sum: go.mod
