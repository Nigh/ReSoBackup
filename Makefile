APP_NAME := rsbackup
DIST_DIR := dist
PKG := .

GO ?= go

.DEFAULT_GOAL := build-all

DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

TARGETS := \
	linux/amd64 \
	linux/arm64 \
	windows/amd64 \
	windows/arm64 \
	darwin/amd64 \
	darwin/arm64

LDFLAGS := -s -w
GOFLAGS := -trimpath -buildvcs=false

.PHONY: build-all clean release-dirs

define artifact_path
$(DIST_DIR)/$(APP_NAME)_$(subst /,_,$(1))$(if $(filter windows/%,$(1)),.exe,)
endef

build-all: clean release-dirs $(foreach target,$(TARGETS),$(call artifact_path,$(target)))

clean:
	rm -rf $(DIST_DIR)

release-dirs:
	mkdir -p $(DIST_DIR)

define build_target
$(call artifact_path,$(1)): | release-dirs
	@echo "==> Building $$@"
	CGO_ENABLED=0 GOOS="$(word 1,$(subst /, ,$(1)))" GOARCH="$(word 2,$(subst /, ,$(1)))" \
	$(GO) build $(GOFLAGS) -ldflags '$(LDFLAGS)' -o "$$@" $(PKG)
endef

$(foreach target,$(TARGETS),$(eval $(call build_target,$(target))))
