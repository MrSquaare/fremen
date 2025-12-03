.PHONY: install build lint lint-fix test clean

GO_BINARY = ./dist/fremen

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
	go test ./tests/... -v

clean:
	rm -f $(GO_BINARY)
