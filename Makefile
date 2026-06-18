BINARY ?= elispvm
BIN_DIR ?= bin
DIST_DIR ?= dist
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
GOOS ?= $(shell if command -v go >/dev/null 2>&1; then go env GOOS; else echo unknown; fi)
GOARCH ?= $(shell if command -v go >/dev/null 2>&1; then go env GOARCH; else echo unknown; fi)
PACKAGE_NAME := vibeEmacsLispVm_$(VERSION)_$(GOOS)_$(GOARCH)
LDFLAGS := -X main.version=$(VERSION)

.PHONY: all fmt test vet build repl package clean

all: test build

fmt:
	gofmt -w .

test:
	go test ./...

vet:
	go vet ./...

build:
	mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY) ./cmd/elispvm

repl: build
	$(BIN_DIR)/$(BINARY)

package: build
	rm -rf $(DIST_DIR)/$(PACKAGE_NAME)
	mkdir -p $(DIST_DIR)/$(PACKAGE_NAME)
	cp $(BIN_DIR)/$(BINARY) $(DIST_DIR)/$(PACKAGE_NAME)/
	cp README.md LICENSE $(DIST_DIR)/$(PACKAGE_NAME)/
	tar -C $(DIST_DIR) -czf $(DIST_DIR)/$(PACKAGE_NAME).tar.gz $(PACKAGE_NAME)

clean:
	rm -rf $(BIN_DIR) $(DIST_DIR) coverage.out
