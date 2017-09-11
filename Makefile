SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

BINARY=runner

VERSION=v0.0.1

.DEFAULT_GOAL: $(BINARY)

$(BINARY): $(SOURCES)
	CGO_ENABLED=0 go build -o ${BINARY} *.go

.PHONY: test
test:
	go test -v -covermode=atomic -coverprofile=covarage.out ./...
	go tool cover -func=covarage.out
