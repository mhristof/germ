#
# vim:ft=make
#

SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.ONESHELL:

GIT_REF := $(shell git rev-parse --short HEAD)
GIT_TAG := $(shell git name-rev --tags --name-only $(GIT_REF))

./bin/germ: $(shell find ./ -name '*.go')
	go build -o bin/germ -ldflags "-X github.com/mhristof/germ/cmd.version=$(GIT_TAG)+$(GIT_REF)" main.go

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
	rm -rf bin/germ

help:           ## Show this help.
	@grep '.*:.*##' Makefile | grep -v grep  | sort | sed 's/:.* ##/:/g' | column -t -s:
.PHONY: help
