GO ?= go
CONFIG_PATH ?= ./dev-config.yml

all: build
.PHONY: all

build:
	mkdir -p build
	$(GO) build -o build

.PHONY: build

run: build
	CONFIG_PATH=$(CONFIG_PATH) ./build/server-tool

.PHONY: run

