
APP_NAME := ota-updater
VERSION := 0.8.0
SERVER_NAME := update-server
BUILD_DIR := ./build
RELEASES_DIR := ./releases

# Build flags
GO := go
LDFLAGS := -ldflags "-X github.com/noamstrauss/ota-updater/version.Version=$(VERSION)"
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
SERVER_BIN := $(SERVER_NAME)$(BIN_EXT)

#.PHONY: all clean build run-app run-server release
.PHONY: all clean build run-app release

all: build

#build: build-app build-server
build: build-app
build-app:
	@echo "Building application..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BIN_NAME) .

#build-server:
#	@echo "Building update server..."
#	@mkdir -p $(BUILD_DIR)
#	$(GO) build -o $(BUILD_DIR)/$(SERVER_BIN) ./server

run-app:
	@echo "Running application..."
#	$(BUILD_DIR)/$(BIN_NAME) --update-url="http://localhost:8080"
	$(BUILD_DIR)/$(BIN_NAME)

#run-server: build-server
#	@echo "Running update server on port 8080..."
#	$(BUILD_DIR)/$(SERVER_BIN) --port=8080 --releases-dir=$(RELEASES_DIR)

# Create a new release
#release: build-app
#	@echo "Creating release $(VERSION) for $(PLATFORM)/$(ARCH)..."
#	@mkdir -p $(RELEASES_DIR)/$(PLATFORM)/$(ARCH)
#	@cp $(BUILD_DIR)/$(BIN_NAME) $(RELEASES_DIR)/$(PLATFORM)/$(ARCH)/$(VERSION).bin
#	@echo "Release created. Upload it to the server using:"
#	@echo "curl -F \"platform=$(PLATFORM)\" -F \"arch=$(ARCH)\" -F \"version=$(VERSION)\" -F \"binary=@$(RELEASES_DIR)/$(PLATFORM)/$(ARCH)/$(VERSION).bin\" http://localhost:8080/upload"

clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)