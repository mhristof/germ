#
# vim:ft=make
#


MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := all
.DELETE_ON_ERROR:
.ONESHELL:

GIT_REF := $(shell git rev-parse --short HEAD)
GIT_TAG := $(shell git name-rev --tags --name-only $(GIT_REF))

all: ./bin/germ.darwin

./bin/germ.%: $(shell find ./ -name '*.go')
	GOOS=$* go build -o $@ -ldflags "-X github.com/mhristof/germ/cmd.version=$(GIT_TAG)+$(GIT_REF)" main.go

gen:
	go run main.go generate -n
.PHONY: gen

write: test
	go run main.go generate -w
.PHONY: write

test:
	go test ./...
.PHONY: test

diff:
	go run main.go generate --diff
.PHONY: diff

clean:
	rm -rf bin/germ.*

help:           ## Show this help.
	@grep '.*:.*##' Makefile | grep -v grep  | sort | sed 's/:.* ##/:/g' | column -t -s:
.PHONY: help
