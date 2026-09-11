BIN     := dist/mcp-wp-go
SRC     := $(shell find . -type f -name '*.go' -not -path './vendor/*') go.mod go.sum
VERSION ?= 0.1.3
PREFIX  ?= /usr/local
LDFLAGS := -s -w -X mcp-wp-go/internal/server.Version=$(VERSION)

.PHONY: all build test fmt vet lint check tidy release-cross install uninstall clean

all: build

build: $(BIN)

$(BIN): $(SRC)
	@mkdir -p dist
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o $@ .
	@echo "  built:      $$(du -h $@ | cut -f1)"
	@if command -v upx >/dev/null; then \
	  if upx -q --best --lzma $@ >/dev/null 2>&1; then \
	    echo "  compressed: $$(du -h $@ | cut -f1)"; \
	  else \
	    echo "  upx skipped; binary left as built"; \
	  fi; \
	else \
	  echo "  upx not installed, binary left uncompressed (sudo apt install upx-ucl)"; \
	fi

test:
	go test -race ./...

fmt:
	@gofmt -w $$(find . -type f -name '*.go' -not -path './vendor/*')

vet:
	go vet ./...

lint: vet
	@test -z "$$(gofmt -l $$(find . -type f -name '*.go' -not -path './vendor/*'))" || { echo "gofmt needed:"; gofmt -l $$(find . -type f -name '*.go' -not -path './vendor/*'); exit 1; }

tidy:
	go mod tidy

check: fmt lint test
	go mod verify

release-cross:
	@mkdir -p dist/pkg
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" \
	  -o dist/pkg/mcp-wp-go .
	tar -czf dist/mcp-wp-go_$(VERSION)_linux_amd64.tar.gz \
	  -C dist/pkg mcp-wp-go -C $(CURDIR) LICENSE README.md
	@echo "  dist/mcp-wp-go_$(VERSION)_linux_amd64.tar.gz"
	@rm -rf dist/pkg

install: build
	install -m 0755 $(BIN) $(PREFIX)/bin/mcp-wp-go

uninstall:
	rm -f $(PREFIX)/bin/mcp-wp-go

clean:
	rm -rf dist
