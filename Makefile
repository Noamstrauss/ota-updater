APP_NAME := ota-updater
VERSION := 0.8.0
BUILD_DIR := ./build

# Platform detection
ifeq ($(OS),Windows_NT)
    PLATFORM := windows
    BIN_EXT := .exe
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Linux)
       PLATFORM := linux
    endif
    ifeq ($(UNAME_S),Darwin)
       PLATFORM := darwin
    endif
    BIN_EXT :=
endif

ARCH := $(shell go env GOARCH)
BIN_NAME := $(APP_NAME)$(BIN_EXT)

.PHONY: all clean build run-app

all: build

build: build-app

build-app:
	@echo "Building application..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "-X github.com/noamstrauss/ota-updater/version.Version=$(VERSION)" -o $(BUILD_DIR)/$(BIN_NAME) .

run-app:
	@echo "Running application..."
	$(BUILD_DIR)/$(BIN_NAME)

clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)