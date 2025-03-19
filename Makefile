APP_NAME := ota-updater
VERSION := 0.2.0
BUILD_DIR := ./build

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

.PHONY: clean build run build-run help release tag

help:
	@echo "Available targets:"
	@echo "  build      - Build the application"
	@echo "  run        - Run the application (build if necessary)"
	@echo "  build-run  - Build and then run the application"
	@echo "  clean      - Remove build artifacts"
	@echo "  release    - Create a release build for current platform"
	@echo "  tag        - Create a git tag for the current version"
	@echo "  release-tag - Create and push a git tag to trigger GitHub workflow"
	@echo "  help       - Show this help message"

build:
	@echo "Building application..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "-X github.com/noamstrauss/ota-updater/version.Version=$(VERSION)" -o $(BUILD_DIR)/$(BIN_NAME) .

run: $(BUILD_DIR)/$(BIN_NAME)
	@echo "Running application..."
	$(BUILD_DIR)/$(BIN_NAME)

$(BUILD_DIR)/$(BIN_NAME):
	@$(MAKE) build

build-run: clean build run

release:
	@echo "Creating release build for $(PLATFORM)/$(ARCH)..."
	@mkdir -p $(BUILD_DIR)/release
	GOOS=$(PLATFORM) GOARCH=$(ARCH) go build -ldflags "-X github.com/noamstrauss/ota-updater/version.Version=$(VERSION) -s -w" -o $(BUILD_DIR)/release/$(APP_NAME)-$(VERSION)-$(PLATFORM)-$(ARCH)$(BIN_EXT) .

tag:
	@echo "Creating git tag $(VERSION)..."
	@git tag -a $(VERSION) -m "$(VERSION)"
	@echo "Tag created locally. To push this tag and trigger a release workflow, run: git push origin $(VERSION)"

release-tag: tag
	@echo "Pushing git tag $(VERSION)..."
	@git push origin $(VERSION)
	@echo "Tag pushed. GitHub Actions workflow should start automatically to build the release."

clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)