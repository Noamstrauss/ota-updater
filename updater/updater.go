// updater/github_updater.go
package updater

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/google/go-github/v40/github"
	"golang.org/x/oauth2"
)

// Config contains the configuration for the updater
type Config struct {
	CurrentVersion string
	GithubRepo     string
	GithubToken    string
	ExecutablePath string
}

// CheckAndUpdate checks for an update and applies it if available
func CheckAndUpdate(config Config) (bool, error) {
	// Parse repository owner and name
	parts := strings.Split(config.GithubRepo, "/")
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid GitHub repo format, should be 'owner/repo'")
	}
	owner, repo := parts[0], parts[1]

	// Setup GitHub client
	ctx := context.Background()
	var client *github.Client

	if config.GithubToken != "" {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: config.GithubToken},
		)
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	} else {
		client = github.NewClient(nil)
	}

	// Get the latest release
	release, _, err := client.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		return false, fmt.Errorf("failed to get latest release: %w", err)
	}

	// Clean version string (remove 'v' prefix if present)
	latestVersion := strings.TrimPrefix(*release.TagName, "v")
	currentVersion := strings.TrimPrefix(config.CurrentVersion, "v")

	// Check if an update is available
	if latestVersion <= currentVersion {
		return false, nil
	}

	log.Printf("Update available: %s", latestVersion)

	// Get current platform info
	platform := runtime.GOOS
	arch := runtime.GOARCH

	// Find the appropriate asset for the current platform and architecture
	assetName := fmt.Sprintf("%s-%s", platform, arch)
	var downloadURL string

	for _, asset := range release.Assets {
		if asset.BrowserDownloadURL != nil && strings.Contains(*asset.Name, assetName) {
			downloadURL = *asset.BrowserDownloadURL
			break
		}
	}

	if downloadURL == "" {
		return false, fmt.Errorf("no suitable asset found for %s/%s", platform, arch)
	}

	// Download and apply the update
	return downloadAndApplyUpdate(config.GithubToken, config.ExecutablePath, downloadURL)
}

// downloadAndApplyUpdate downloads and applies the update
func downloadAndApplyUpdate(token, executablePath, downloadURL string) (bool, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	// Create the request
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return false, err
	}

	// Add headers
	req.Header.Set("User-Agent", "ota-updater-client")
	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("download failed with status code %d", resp.StatusCode)
	}

	// Create temporary file for the download
	tmpDir := os.TempDir()
	tempFile, err := os.CreateTemp(tmpDir, "update_*.bin")
	if err != nil {
		return false, fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath) // Clean up temp file on function exit

	// Copy the downloaded binary to the temp file
	_, err = io.Copy(tempFile, resp.Body)
	tempFile.Close()
	if err != nil {
		return false, fmt.Errorf("failed to write downloaded file: %w", err)
	}

	// Set executable permissions
	if err := os.Chmod(tempPath, 0755); err != nil {
		return false, fmt.Errorf("failed to set permissions: %w", err)
	}

	// Create backup of current executable
	backupPath := executablePath + ".bak"
	if err := copyFile(executablePath, backupPath); err != nil {
		return false, fmt.Errorf("failed to create backup: %w", err)
	}

	// Replace the executable
	if runtime.GOOS == "windows" {
		// On Windows, we need to use a batch file for replacement
		return true, replaceExecutableWindows(tempPath, executablePath)
	}

	// On Unix-like systems, we can replace directly
	if err := os.Rename(tempPath, executablePath); err != nil {
		// Try to restore backup
		os.Rename(backupPath, executablePath)
		return false, fmt.Errorf("failed to replace executable: %w", err)
	}

	return true, nil
}

// replaceExecutableWindows creates a batch file to replace the executable after process exit
func replaceExecutableWindows(newFile, targetFile string) error {
	batchContent := fmt.Sprintf(`@echo off
:retry
ping -n 2 127.0.0.1 > nul
del "%s"
if exist "%s" goto retry
copy /y "%s" "%s"
start "" "%s" %s
del "%s"
del "%%~f0"
`, targetFile, targetFile, newFile, targetFile, targetFile, strings.Join(os.Args[1:], " "), newFile)

	batchPath := filepath.Join(os.TempDir(), "update.bat")
	if err := os.WriteFile(batchPath, []byte(batchContent), 0700); err != nil {
		return err
	}

	cmd := exec.Command("cmd", "/c", "start", "/b", batchPath)
	return cmd.Start()
}

// RestartApplication restarts the application
func RestartApplication(executablePath string, args []string) {
	cmd := exec.Command(executablePath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		log.Printf("Failed to restart application: %v", err)
		return
	}

	os.Exit(0)
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}
