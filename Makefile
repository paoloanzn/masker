APP_NAME := masker
CMD_PATH := ./cmd/masker
APP_DISPLAY_NAME := Masker
BUILD_DIR := .build
APP_BUNDLE := $(BUILD_DIR)/$(APP_DISPLAY_NAME).app
APP_CONTENTS := $(APP_BUNDLE)/Contents
APP_MACOS := $(APP_CONTENTS)/MacOS

.PHONY: build run test fmt clean

build:
	go build -o $(APP_NAME) $(CMD_PATH)

run:
	@if [ "$$(uname -s)" = "Darwin" ]; then \
		mkdir -p "$(APP_MACOS)"; \
		cp build/macos/Info.plist "$(APP_CONTENTS)/Info.plist"; \
		go build -o "$(APP_MACOS)/$(APP_NAME)" $(CMD_PATH); \
		open "$(APP_BUNDLE)"; \
	else \
		go run $(CMD_PATH); \
	fi

test:
	go test ./...

fmt:
	gofmt -w $(shell find cmd internal -name '*.go' -type f)

clean:
	rm -f $(APP_NAME)
	rm -rf $(BUILD_DIR)
