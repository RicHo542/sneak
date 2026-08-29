.PHONY: build build-all build-macos build-linux build-windows install-mac install-wsl test clean

BINARY 	:= sneak
VERSION := 0.0.1
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -ldflags="-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

PLATFORMS_DARWIN  := darwin/amd64 darwin/arm64
PLATFORMS_LINUX   := linux/amd64 linux/arm64
PLATFORMS_WINDOWS := windows/amd64

define build_platform
	set -e; rm -rf bin/pkg; mkdir -p bin/pkg; \
	for platform in $(1); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		ext=""; [ $$os = windows ] && ext=".exe"; \
		name=$(BINARY)-$${os}-$${arch}; \
		archive=$(BINARY)-$(VERSION)-$${os}-$${arch}; \
		echo "building $${name}..."; \
		GOOS=$${os} GOARCH=$${arch} CGO_ENABLED=0 \
			go build $(LDFLAGS) -o bin/$${name}$${ext} ./cmd/sneak; \
		cp bin/$${name}$${ext} bin/pkg/sneak$${ext}; \
		echo "creating $${archive}.tar.gz..."; \
		tar -czf bin/$${archive}.tar.gz -C bin/pkg sneak$${ext}; \
		rm bin/pkg/sneak$${ext}; \
	done; \
	rm -rf bin/pkg
endef

build:
	go build $(LDFLAGS) -o bin/$(BINARY) ./cmd/sneak

build-all: build-macos build-linux build-windows

build-macos:
	@$(call build_platform,$(PLATFORMS_DARWIN))

build-linux:
	@$(call build_platform,$(PLATFORMS_LINUX))

build-windows:
	@$(call build_platform,$(PLATFORMS_WINDOWS))

install-wsl: build-linux
	@echo "Installing sneak for WSL..."
	@sudo mkdir -p /usr/local/bin
	@tar -xzf bin/$(BINARY)-$(VERSION)-linux-amd64.tar.gz -C /tmp
	@sudo mv /tmp/$(BINARY) /usr/local/bin/$(BINARY)
	@sudo chmod +x /usr/local/bin/$(BINARY)
	@hash -r
	@echo "Installed $$(sneak version 2>/dev/null || echo $(BINARY)) to /usr/local/bin/$(BINARY)"

install-macos: build-macos
	@echo "Installing sneak for macOS..."
	@sudo mkdir -p /usr/local/bin
	@tar -xzf bin/$(BINARY)-$(VERSION)-darwin-arm64.tar.gz -C /tmp
	@sudo mv /tmp/$(BINARY) /usr/local/bin/$(BINARY)
	@sudo chmod +x /usr/local/bin/$(BINARY)
	@echo "Installed $$(sneak version 2>/dev/null || echo $(BINARY)) to /usr/local/bin/$(BINARY)"

test:
	go test ./tests

clean:
	rm -rf bin/*
