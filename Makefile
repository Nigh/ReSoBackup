APP_NAME := rsbackup
DIST_DIR := dist
GO ?= go
GOEXE := $(shell $(GO) env GOEXE)
GOFLAGS := -trimpath -buildvcs=false

.DEFAULT_GOAL := build
.PHONY: build build-all cli dev package test clean

build:
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) -ldflags '-s -w' -o $(DIST_DIR)/$(APP_NAME)$(GOEXE) .

build-all:
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags '-s -w' -o $(DIST_DIR)/$(APP_NAME)_linux_amd64 .
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build $(GOFLAGS) -ldflags '-s -w' -o $(DIST_DIR)/$(APP_NAME)_windows_amd64.exe .
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GO) build $(GOFLAGS) -ldflags '-s -w' -o $(DIST_DIR)/$(APP_NAME)_darwin_arm64 .

cli:
	mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 $(GO) build $(GOFLAGS) -tags cli -ldflags '-s -w' -o $(DIST_DIR)/$(APP_NAME)-cli$(GOEXE) .

dev:
	$(GO) tool mygo dev

package:
	$(GO) tool mygo build

test:
	$(GO) vet ./...
	$(GO) test ./...
	$(GO) test -tags cli ./...

clean:
	rm -rf $(DIST_DIR)
