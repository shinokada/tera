.PHONY: build test clean run

build:
	go build -o tera cmd/tera/main.go

test:
	go test -v ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

run:
	go run cmd/tera/main.go

clean:
	rm -f tera coverage.out

install:
	go install cmd/tera/main.go