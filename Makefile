.DEFAULT_GOAL := all
SHELL := /bin/bash -Eeuo pipefail

.PHONY: all
all: setup gen

.PHONY: setup
setup:
	gex --build

.PHONY: gen
gen:
	@go generate ./...

.PHONY: test
test:
	@go test -v ./...