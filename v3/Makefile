.PHONY: build test clean clean-cache clean-all run lint lint-fix coverage install

build: test
	go build -o tera cmd/tera/main.go

test:
	go test -v ./...

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

run:
	go run cmd/tera/main.go

# Clean build artifacts
clean:
	rm -f tera coverage.out

# Clean Go caches (test cache and build cache)
clean-cache:
	go clean -testcache -cache

# Clean everything (build artifacts + caches)
clean-all: clean clean-cache

install:
	go install cmd/tera/main.go