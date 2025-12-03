.PHONY: install build format format-check lint test clean

GO_BINARY = ./dist/fremen

install:
	go mod download

build: 
	mkdir -p ./dist
	go build -o $(GO_BINARY) ./cmd/fremen

format:
	golangci-lint fmt ./...

format-check:
	@test -z $$(golangci-lint fmt -d ./...)

lint:
	golangci-lint run ./...

test:
	go test ./tests/... -v

clean:
	rm -f $(GO_BINARY)
