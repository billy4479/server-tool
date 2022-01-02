GO ?= go
BASE_DIR ?= ./test-dir

all: build
.PHONY: all

build:
	mkdir -p build
	$(GO) build -o build

.PHONY: build

run: build
	BASE_DIR=$(BASE_DIR) ./build/server-tool

.PHONY: run

