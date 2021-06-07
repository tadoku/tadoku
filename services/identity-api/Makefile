.DEFAULT_GOAL := all
SHELL := /bin/bash -Eeuo pipefail

.PHONY: all
all: setup gen lint test run

.PHONY: setup
setup:
	gex --build

.PHONY: gen
gen:
	@go generate ./...

.PHONY: lint
lint:
	gex golint -set_exit_status ./...
	gex ineffassign .

.PHONY: test
test:
	@make lint
	@go test -v ./... -count=1

.PHONY: run
run:
	@go run cmd/server/main.go
