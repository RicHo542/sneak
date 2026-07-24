BINARY := sneak

.PHONY: build test clean

build:
	go build -o bin/$(BINARY) ./cmd/sneak

test:
	go test ./...

clean:
	rm -rf bin
