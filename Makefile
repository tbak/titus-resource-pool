# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

GOCMD  ?= GO111MODULE=on go
GOLANGCI_VERSION ?= v1.50.1

GOBIN_TOOL := $(shell which gobin || echo $(GOBIN)/gobin)

# Run go fmt against code
.PHONY: fmt
fmt: $(GOBIN_TOOL)
	$(GOBIN_TOOL) -run golang.org/x/tools/cmd/goimports -w $(shell go list -f '{{.Dir}}' ./...)

# Run lint against code
.PHONY: lint
lint: $(GOBIN)/golangci-lint
	$(GOCMD) mod tidy
	$(GOBIN)/golangci-lint run

$(GOBIN)/golangci-lint: $(GOBIN)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) $(GOLANGCI_VERSION)

# Run tests
test:
	go test ./... -coverprofile cover.out
