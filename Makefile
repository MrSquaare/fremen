.PHONY: install build format format-check test clean

GO_BINARY = ./dist/fremen

install:
	go mod download

build: 
	mkdir -p ./dist
	go build -o $(GO_BINARY) ./cmd/fremen

format:
	go fmt ./...

format-check:
	@test -z $$(gofmt -l .)

test:
	go test ./tests/... -v

clean:
	rm -f $(GO_BINARY)
