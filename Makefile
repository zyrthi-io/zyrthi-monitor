# zyrthi-monitor Makefile

BINARY = zyrthi-monitor

.PHONY: all build test clean install

all: build

build:
	go build -o $(BINARY) ./cmd

test:
	go test -v -race -coverprofile=coverage.out ./...

clean:
	rm -f $(BINARY) coverage.out

install: build
	go install ./cmd
