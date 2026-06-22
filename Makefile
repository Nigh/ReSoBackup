APP_NAME := rsbackup
DIST_DIR := dist
PKG := .

GO ?= go

.DEFAULT_GOAL := build-all

DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

ifeq ($(shell go env GOOS),windows)
    LDFLAGS := -s -w -H windowsgui
else
    LDFLAGS := -s -w
endif
GOFLAGS := -trimpath -buildvcs=false

.PHONY: build-all clean release-dirs

$(DIST_DIR)/$(APP_NAME): | release-dirs
	@echo "==> Building $@"
	CGO_ENABLED=1 $(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o "$@" $(PKG)

build-all: clean release-dirs $(DIST_DIR)/$(APP_NAME)

clean:
	rm -rf $(DIST_DIR)

release-dirs:
	mkdir -p $(DIST_DIR)
