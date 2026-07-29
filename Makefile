.PHONY: build test clean

BINARY 	:= sneak
VERSION := 0.0.1
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -ldflags="-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"


build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/sneak

test:
	go test ./...

clean:
	rm -rf bin
