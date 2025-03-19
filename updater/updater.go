// updater/updater.go
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

// Config contains the config for the updater
type Config struct {
	CurrentVersion string
	GithubRepo     string
	GithubToken    string
	ExecutablePath string
}

// CheckAndUpdate checks for an update and applies it if available
func CheckAndUpdate(config Config) (bool, error) {
	parts := strings.Split(config.GithubRepo, "/")
	if len(parts) != 2 {
		return false, fmt.Errorf("invalid GitHub repo format, shoulf be 'owner/repo'")
	}
	owner, repo := parts[0], parts[1]

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

	release, _, err := client.Repositories.GetLatestRelease(ctx, owner, repo)
	if err != nil {
		return false, fmt.Errorf("failed to get latest release: %w", err)
	}

	latestVersion := *release.TagName

	if strings.HasPrefix(latestVersion, "v") {
		latestVersion = strings.TrimPrefix(latestVersion, "v")
	}

	currentVersion := config.CurrentVersion

	if latestVersion <= currentVersion {
		return false, nil
	}

	log.Printf("Update available: %s", latestVersion)

	platform := runtime.GOOS
	arch := runtime.GOARCH

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

	return downloadAndApplyUpdate(config.GithubToken, config.ExecutablePath, downloadURL)
}

// downloadAndApplyUpdate downloads and applies the update
func downloadAndApplyUpdate(token, executablePath, downloadURL string) (bool, error) {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return false, err
	}

	if token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("download failed with status code %d", resp.StatusCode)
	}

	tmpDir := os.TempDir()
	tempFile, err := os.CreateTemp(tmpDir, "update_*.bin")
	if err != nil {
		return false, fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	_, err = io.Copy(tempFile, resp.Body)
	tempFile.Close()
	if err != nil {
		return false, fmt.Errorf("failed to write downloaded file: %w", err)
	}

	if err := os.Chmod(tempPath, 0755); err != nil {
		return false, fmt.Errorf("failed to set permissions: %w", err)
	}

	backupPath := executablePath + ".bak"
	if err := copyFile(executablePath, backupPath); err != nil {
		return false, fmt.Errorf("failed to create backup: %w", err)
	}

	if runtime.GOOS == "windows" {
		// On Windows, we need to use a batch file for replacement
		return true, replaceExecutableWindows(tempPath, executablePath)
	}

	// On not windows replace directly
	if err := os.Rename(tempPath, executablePath); err != nil {
		// If failed restore backup
		os.Rename(backupPath, executablePath)
		return false, fmt.Errorf("failed to replace executable: %w", err)
	}

	return true, nil
}

// replaceExecutableWindows creates a batch file for Windows to replace the executable after process exit
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
