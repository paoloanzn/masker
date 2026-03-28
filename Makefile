APP_NAME := masker
CMD_PATH := ./cmd/masker
APP_DISPLAY_NAME := Masker
BUILD_DIR := .build
APP_BUNDLE := $(BUILD_DIR)/$(APP_DISPLAY_NAME).app
APP_CONTENTS := $(APP_BUNDLE)/Contents
APP_MACOS := $(APP_CONTENTS)/MacOS
INSTALL_DIR ?= $(shell \
	os="$$(uname -s)"; \
	case "$$os" in \
		Darwin) candidates="$$XDG_BIN_HOME $$HOME/.local/bin $$HOME/bin /opt/homebrew/bin /usr/local/bin" ;; \
		*) candidates="$$XDG_BIN_HOME $$HOME/.local/bin $$HOME/bin /usr/local/bin" ;; \
	esac; \
	for dir in $$candidates; do \
		if [ -n "$$dir" ] && [ -d "$$dir" ] && [ -w "$$dir" ]; then \
			printf '%s' "$$dir"; \
			exit 0; \
		fi; \
	done; \
	printf '%s' "$$HOME/.local/bin")

.PHONY: build run install test fmt clean

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

install: build
	@mkdir -p "$(INSTALL_DIR)"
	@install -m 0755 "$(APP_NAME)" "$(INSTALL_DIR)/$(APP_NAME)"
	@printf 'Installed %s to %s\n' "$(APP_NAME)" "$(INSTALL_DIR)/$(APP_NAME)"

test:
	go test ./...

fmt:
	gofmt -w $(shell find cmd internal -name '*.go' -type f)

clean:
	rm -f $(APP_NAME)
	rm -rf $(BUILD_DIR)
