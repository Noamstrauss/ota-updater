
# OTA Updater

![Release](https://github.com/noamstrauss/ota-updater/actions/workflows/release.yaml/badge.svg)

The OTA Updater is a self-updating program designed to automatically check for and apply updates to itself. It supports multiple platforms (Windows, macOS, Linux) and can fetch updates directly from GitHub repositories.

## Table of Contents

- [Features](#features)
- [Pre-requisites](#Pre-requisites)
- [How It Works](#how-it-works)
- [Quick Start](#quick-start)
- [Usage](#usage)
- - [Updating the Version](#updating-the-version)
- - [Makefile Commands](#makefile-commands)
- [Configuration](#configuration)
- - [Default Configuration](#default-configuration)
- - [Environment Variables](#environment-variables)
- [Improvements](#improvements)
- [Contributing](#contributing)
- [License](#license)

## Features

* Automatic Updates: The program checks for updates at a configurable interval and downloads new releases if available.
* Cross-Platform Support: Works on Windows, macOS, and Linux.
* GitHub Integration: Fetches the latest releases from a specified GitHub repository.
* Backup & Rollback: Creates backups of the current executable before applying updates, with an automatic rollback mechanism in case of failure.
* Configurable Update Interval: Configure how often the program checks for updates.

## Pre-requisites

Before using this OTA Updater, ensure that you have the following:

- [Go](https://go.dev/doc/install) (v1.24 or later)
- [Make](https://formulae.brew.sh/formula/make)
- [Git](https://git-scm.com/downloads)
- **GitHub token** if the repository is private. This is necessary to authenticate and access private repositories when checking for updates.

## How It Works

1. Reads the current application version.
2. Fetches the latest release from GitHub.
3. Compares versions and downloads the update if a newer release exists.
4. Replaces the existing executable with the updated version.

## Quick Start
**Clone the repository:**

```bash
git clone https://github.com/noamstrauss/ota-updater.git
cd ota-updater
```

**Build and run the application:**

```bash
make build-run
```

## Usage

### Updating the Version

To update the OTA Updater version, modify the `VERSION` variable in the `Makefile`:

```bash
VERSION := 0.2.0
```

Then, either push the version tag locally and to GitHub:

```bash
git tag 0.2.0 && git push origin 0.2.0
```

Or use the release-tag command to automate tagging and pushing:

```bash
make release-tag
```

### Makefile Commands

```bash
make build - Compiles the application.

make run - Runs the application (builds if necessary).

make build-run - Builds and runs the application.

make clean - Removes build artifacts.

make release - Creates a release build for the current platform.

make tag - Creates a Git tag for the current version.

make release-tag - Creates and pushes a Git tag to trigger GitHub Actions.
```

## Configuration

OTA Updater uses a JSON configuration file to store settings. The default config is created at runtime if missing.

### Default Configuration

```json
{
  "update_interval": 60,
  "github_repo": "noamstrauss/ota-updater",
  "log_level": "info"
}
```

### Environment Variables

You can override settings with environment variables:

* `GITHUB_TOKEN` - (Optional) Personal access token for private repositories.

* `GITHUB_REPO` - Overrides the repository to check for updates.

* `LOG_LEVEL` - Logging verbosity (debug, info, warn, error).

* `UPDATE_INTERVAL` - Update check interval in minutes.

## Improvements

Possible improvements that can be made

1. Validate Version Using Checksum
2. Add tests

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Feel free to submit issues and pull requests.