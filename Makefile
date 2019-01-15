.DEFAULT_GOAL := all
SHELL := /bin/bash -Eeuo pipefail

.PHONY: all
all: setup gen lint migrate test

.PHONY: setup
setup:
	gex --build

.PHONY: gen
gen:
	@go generate ./...

.PHONY: lint
lint:
	gex golint ./...

.PHONY: migrate
migrate:
	@go run cmd/migrate/main.go

.PHONY: test
test:
	@go test -v ./...