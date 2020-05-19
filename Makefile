#
# vim:ft=make
#

SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.ONESHELL:

./bin/germ: $(shell find ./ -name '*.go')
	go build -o bin/germ main.go

gen:
	go run main.go generate -n
.PHONY: gen

write: test
	go run main.go generate -w
.PHONY: write

test:
	go test ./...
.PHONY: test

help:           ## Show this help.
	@grep '.*:.*##' Makefile | grep -v grep  | sort | sed 's/:.* ##/:/g' | column -t -s:
.PHONY: help
