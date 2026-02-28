SHELL := /usr/bin/env bash

GO ?= go
GOFMT ?= gofmt
GIT_COMMIT := $(shell git rev-parse --short=8 HEAD 2>/dev/null || echo dev)
BUILD_TS := $(shell date +%s)
LDFLAGS := -X owlx/internal/build.Commit=$(GIT_COMMIT) -X owlx/internal/build.Timestamp=$(BUILD_TS)
HOST_OS := $(shell uname -s | tr '[:upper:]' '[:lower:]')
HOST_ARCH_RAW := $(shell uname -m)
ifeq ($(HOST_ARCH_RAW),x86_64)
HOST_ARCH := amd64
else ifeq ($(HOST_ARCH_RAW),aarch64)
HOST_ARCH := arm64
else
HOST_ARCH := $(HOST_ARCH_RAW)
endif
TARGET_OS := $(if $(GOOS),$(GOOS),$(HOST_OS))
TARGET_ARCH_RAW := $(if $(GOARCH),$(GOARCH),$(HOST_ARCH))
ifeq ($(TARGET_ARCH_RAW),x86_64)
TARGET_ARCH := amd64
else ifeq ($(TARGET_ARCH_RAW),aarch64)
TARGET_ARCH := arm64
else
TARGET_ARCH := $(TARGET_ARCH_RAW)
endif
TMUX_PLATFORM := $(TARGET_OS)_$(TARGET_ARCH)

BUILD_OUT ?= owlx
TMUX_STATIC_HOME ?= $(HOME)/.cache/owlx/tmux-static
TMUX_STRIPPED ?= $(TMUX_STATIC_HOME)/bin/tmux.$(TARGET_OS)-$(TARGET_ARCH).stripped.gz
TMUX_META_FILE ?= assets/tmux/buildinfo.env
TMUX_EMBEDDED ?= assets/tmux/$(TMUX_PLATFORM)/tmux
CLEAN_BUILD ?= 0

TMUX_STRIPPED_EXISTS := $(wildcard $(TMUX_STRIPPED))
GOFILES := $(shell rg --files -g '*.go' 2>/dev/null || find . -type f -name '*.go')

.PHONY: build test lint lint-sh lint-md lint-go gofmt govet shellcheck tmux-update tmux-verify clean

build: tmux-update

build:
	$(GO) build -ldflags "$(LDFLAGS)" -o $(BUILD_OUT) ./cmd/owlx

test: tmux-update
	$(GO) test ./...

lint: lint-sh lint-md lint-go

lint-sh: shellcheck

shellcheck:
	scripts/shellcheck.sh

lint-md:
	scripts/lint-md.sh

lint-go: gofmt govet

gofmt:
	@test -z "$$($(GOFMT) -l $(GOFILES))" || { echo "gofmt required:"; $(GOFMT) -l $(GOFILES); exit 1; }

govet:
	$(GO) vet ./...

$(TMUX_EMBEDDED): scripts/build-embedded-tmux.sh $(TMUX_META_FILE)
	TMUX_STRIPPED="$(TMUX_STRIPPED_EXISTS)" \
	TMUX_TARGET_OS="$(TARGET_OS)" \
	TMUX_TARGET_ARCH="$(TARGET_ARCH)" \
	CLEAN_BUILD="$(CLEAN_BUILD)" \
	TMUX_META_FILE="$(TMUX_META_FILE)" \
	scripts/build-embedded-tmux.sh

tmux-update: $(TMUX_EMBEDDED)

tmux-verify: $(TMUX_EMBEDDED) $(TMUX_META_FILE)
	test -f $(TMUX_EMBEDDED)
	file $(TMUX_EMBEDDED) | grep -Eq "ELF 64-bit|Mach-O 64-bit executable"
	@if [ "$(TARGET_OS)" = "linux" ]; then \
		file $(TMUX_EMBEDDED) | grep -q "statically linked"; \
	fi
	@conf=$$(awk -F= '/^TMUX_DEFAULT_CONF=/{print $$2}' $(TMUX_META_FILE)); \
	sock=$$(awk -F= '/^TMUX_DEFAULT_SOCK=/{print $$2}' $(TMUX_META_FILE)); \
	if [ -z "$$conf" ] || [ -z "$$sock" ]; then \
		echo "missing TMUX_DEFAULT_CONF or TMUX_DEFAULT_SOCK in $(TMUX_META_FILE)"; \
		exit 1; \
	fi; \
	strings -a $(TMUX_EMBEDDED) | grep -F "$$conf" >/dev/null; \
	strings -a $(TMUX_EMBEDDED) | grep -F "$$sock" >/dev/null

clean:
	rm -f $(BUILD_OUT)
	rm -f $(TMUX_EMBEDDED)
	rm -rf $(TMUX_STATIC_HOME)
