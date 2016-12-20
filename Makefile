SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

VERSION=v0.0.1

.DEFAULT_GOAL: test

.PHONY: test
test:
	go test -v -race -covermode=atomic -coverprofile=covarage.out ./...
	go tool cover -func=covarage.out
