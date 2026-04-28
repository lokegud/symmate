.PHONY: build test lint fmt clean

build:
	@go build -ldflags="-X github.com/loke/symmate/pkg/version.Version=dev" -o bin/symmated ./cmd/symmated

test:
	@go test -v ./...

lint:
	@golangci-lint run

fmt:
	@go fmt ./...
	@gofmt -l -w .

clean:
	@rm -rf bin/
	@go clean -testcache
