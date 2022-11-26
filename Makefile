GO ?= go
CONFIG_PATH ?= ./dev-config.yml
VERSION ?= $(shell git describe --always --tags)
LDFLAGS ?= 
RELEASE_LDFLAGS ?= -s -w -X "github.com/billy4479/server-tool.Version=$(VERSION)" $(LDFLAGS)
OUTPUT_DIR ?= ./build


all: build
.PHONY: all

build:
	mkdir -p $(OUTPUT_DIR)
	cd cmd && \
	$(GO) build -ldflags '$(LDFLAGS)' -o .$(OUTPUT_DIR)/server-tool

.PHONY: build

run: build
	CONFIG_PATH=$(CONFIG_PATH) ./$(OUTPUT_DIR)/server-tool

.PHONY: run

build-release:
	cd cmd; \
	GOOS=linux GOARCH=amd64 $(GO) build -ldflags '$(RELEASE_LDFLAGS)' -o .$(OUTPUT_DIR)/server-tool.linux; \
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags '$(RELEASE_LDFLAGS)' -o .$(OUTPUT_DIR)/server-tool.windows.exe

.PHONY: build-release
