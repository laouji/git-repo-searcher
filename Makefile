BINARY_NAME=git-repo-searcher
VERSION=$(shell git rev-parse --short HEAD)

all: test build
build:
	go build -buildvcs=false -o ${BINARY_NAME} ./
gen:
	mockgen -destination=pkg/github/mocks/client.go -source=pkg/github/client.go
test:
	go test ./...
clean:
	go clean
deps:
	go mod vendor
	go mod tidy
up:
	docker compose -p ${BINARY_NAME} up
