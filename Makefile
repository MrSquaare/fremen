.PHONY: install build lint lint-fix test test-coverage clean

DIST_DIR = ./dist
GO_BINARY = $(DIST_DIR)/fremen
COVER_DIR = ./coverdata
COVER_OUT = coverage.out

install:
	go mod download

build: 
	mkdir -p ./dist
	go build -o $(GO_BINARY) ./cmd/fremen

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
	rm -f $(GO_BINARY)
