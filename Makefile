GO ?= go
CONFIG_PATH ?= ./dev-config.yml
VERSION ?= $(shell git describe --always --tags)
LDFLAGS ?= -X "github.com/billy4479/server-tool/lib.Version=$(VERSION)"
RELEASE_LDFLAGS ?= -s -w $(LDFLAGS)
OUTPUT_DIR ?= ./build
ARGS ?= 


all: build
.PHONY: all

build:
	mkdir -p $(OUTPUT_DIR)
	$(GO) build -ldflags '$(LDFLAGS)' -o $(OUTPUT_DIR)

.PHONY: build

run: build
	CONFIG_PATH=$(CONFIG_PATH) $(OUTPUT_DIR)/server-tool $(ARGS)

.PHONY: run

build-release:
	GOOS=linux GOARCH=amd64 $(GO) build -ldflags '$(RELEASE_LDFLAGS)' -o $(OUTPUT_DIR)/server-tool.linux
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags '$(RELEASE_LDFLAGS) -H=windowsgui' -o $(OUTPUT_DIR)/server-tool.windows.exe

install: build-release
	install -Dm755 $(OUTPUT_DIR)/server-tool.linux $(GOPATH)/bin/server-tool

.PHONY: build-release
