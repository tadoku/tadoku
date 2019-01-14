.DEFAULT_GOAL := all
SHELL := /bin/bash -Eeuo pipefail

.PHONY: all
all: setup gen migrate test

.PHONY: setup
setup:
	gex --build

.PHONY: gen
gen:
	@go generate ./...

.PHONY: migrate
migrate:
	@go run cmd/migrate/main.go

.PHONY: test
test:
	@go test -v ./...