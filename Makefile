GO ?= go
CONFIG_PATH ?= ./dev-config.yml
RELEASE_LDFLAGS ?= "-s -w"
OUTPUT_DIR ?= ./build

all: build
.PHONY: all

build:
	mkdir -p $(OUTPUT_DIR)
	$(GO) build -o $(OUTPUT_DIR)

.PHONY: build

run: build
	CONFIG_PATH=$(CONFIG_PATH) ./$(OUTPUT_DIR)/server-tool

.PHONY: run

build-release:
	GOOS=linux GOARCH=amd64 $(GO) build -ldflags $(RELEASE_LDFLAGS) -o $(OUTPUT_DIR)/server-tool.linux
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags $(RELEASE_LDFLAGS) -o $(OUTPUT_DIR)/server-tool.windows.exe

.PHONY: build-release
