APP_NAME := masker
CMD_PATH := ./cmd/masker

.PHONY: build run test fmt clean

build:
	go build -o $(APP_NAME) $(CMD_PATH)

run:
	go run $(CMD_PATH)

test:
	go test ./...

fmt:
	gofmt -w $(shell find cmd internal -name '*.go' -type f)

clean:
	rm -f $(APP_NAME)
