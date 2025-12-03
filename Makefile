.PHONY: install build lint lint-fix test test-coverage clean

GO_SOURCE = ./cmd/fremen
DIST_DIR = ./dist
GO_BINARY = $(DIST_DIR)/fremen
COVER_DIR = ./coverdata
COVER_OUT = coverage.out

install:
	go mod download

build:
	rm -rf $(DIST_DIR)
	mkdir -p $(DIST_DIR)
	go build -o $(GO_BINARY) $(GO_SOURCE)

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run ./... --fix

test:
	go test ./tests/... -v -count=1

test-coverage:
	rm -rf $(COVER_DIR) $(COVER_OUT)
	mkdir -p $(COVER_DIR)
	GOCOVERDIR=$(abspath $(COVER_DIR)) go test ./tests/... -v -count=1
	go tool covdata textfmt -i=$(COVER_DIR) -o=$(COVER_OUT)
	go tool cover -func=$(COVER_OUT)

clean:
	rm -rf $(DIST_DIR) $(COVER_DIR) $(COVER_OUT)

all: install build
