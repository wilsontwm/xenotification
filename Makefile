export GO111MODULE=on
BINARY_NAME=xenotification

all: test build
test:
		godotenv -f .env.dev go test -v ./...
build:
		GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(BINARY_NAME) .
