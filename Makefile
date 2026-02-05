SHELL := /usr/bin/env bash

GO ?= go
GOFMT ?= gofmt
GIT_COMMIT := $(shell git rev-parse --short=8 HEAD 2>/dev/null || echo dev)
BUILD_TS := $(shell date +%s)
LDFLAGS := -X owlx/internal/build.Commit=$(GIT_COMMIT) -X owlx/internal/build.Timestamp=$(BUILD_TS)

BUILD_OUT ?= owlx
TMUX_STATIC_HOME ?= $(HOME)/.cache/owlx/tmux-static
TMUX_STRIPPED ?= $(TMUX_STATIC_HOME)/bin/tmux.linux-amd64.stripped.gz
TMUX_META_FILE ?= assets/tmux/buildinfo.env
TMUX_EMBEDDED ?= assets/tmux/linux_amd64/tmux
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
	CLEAN_BUILD="$(CLEAN_BUILD)" \
	TMUX_META_FILE="$(TMUX_META_FILE)" \
	scripts/build-embedded-tmux.sh

tmux-update: $(TMUX_EMBEDDED)

tmux-verify: $(TMUX_EMBEDDED) $(TMUX_META_FILE)
	test -f $(TMUX_EMBEDDED)
	file $(TMUX_EMBEDDED) | grep -q "ELF 64-bit"
	file $(TMUX_EMBEDDED) | grep -q "statically linked"
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
