.PHONY: build test clean run lint

build:
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

clean:
	rm -f tera coverage.out

install:
	go install cmd/tera/main.go