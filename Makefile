export GO111MODULE=on
BINARY_NAME=xenotification

all: test build
test:
		go test -v ./...
build:
		GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(BINARY_NAME) .
