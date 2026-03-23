.PHONY: build test vet lint install clean

BINARY := hubstaff-tui
VERSION := $(shell cat VERSION 2>/dev/null || echo "dev")

build:
	CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=$(VERSION)" -o $(BINARY) ./cmd/hubstaff-tui

test:
	go test -race ./...

vet:
	go vet ./...

lint:
	golangci-lint run

install: build
	install -m 755 $(BINARY) /usr/local/bin/$(BINARY)
	@echo "Installed $(BINARY) to /usr/local/bin/$(BINARY)"

clean:
	rm -f $(BINARY)

all: vet test build
