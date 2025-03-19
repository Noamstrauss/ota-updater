// updater/updater.go

package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
)

// Config contains the configuration for the updater
type Config struct {
	CurrentVersion string
	UpdateURL      string
	ExecutablePath string
}

// VersionInfo represents version information returned by the update server
type VersionInfo struct {
	Version     string `json:"version"`
	ReleaseDate string `json:"release_date"`
	DownloadURL string `json:"download_url"`
	Checksum    string `json:"checksum"`
}

// CheckAndUpdate checks for an update and applies it if available
func CheckAndUpdate(config Config) (bool, error) {
	// Get current platform info
	platform := runtime.GOOS
	arch := runtime.GOARCH

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Check for updates
	checkURL := fmt.Sprintf("%s/check/%s/%s/%s",
		config.UpdateURL, platform, arch, config.CurrentVersion)

	resp, err := client.Get(checkURL)
	if err != nil {
		return false, fmt.Errorf("failed to check for updates: %w", err)
	}
	defer resp.Body.Close()

	// No update available
	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	// Handle other errors
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("update server returned status code %d", resp.StatusCode)
	}

	// Parse version info
	var versionInfo VersionInfo
	if err := json.NewDecoder(resp.Body).Decode(&versionInfo); err != nil {
		return false, fmt.Errorf("failed to decode version info: %w", err)
	}

	log.Printf("Update available: %s", versionInfo.Version)

	// Download the new version
	return downloadAndApplyUpdate(client, config.ExecutablePath, versionInfo, config.UpdateURL) // Pass UpdateURL
}

// downloadAndApplyUpdate downloads and applies the update
func downloadAndApplyUpdate(client *http.Client, executablePath string, versionInfo VersionInfo, baseURL string) (bool, error) {
	// Handle relative URLs
	downloadURL := versionInfo.DownloadURL
	if !strings.HasPrefix(downloadURL, "http") {
		downloadURL = baseURL + downloadURL
	}

	// Download the binary
	resp, err := client.Get(downloadURL)
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

	// Verify checksum
	if err := verifyChecksum(tempPath, versionInfo.Checksum); err != nil {
		return false, fmt.Errorf("checksum verification failed: %w", err)
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

// verifyChecksum verifies the SHA-256 checksum of a file
func verifyChecksum(filePath, expectedChecksum string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return err
	}

	actualChecksum := hex.EncodeToString(hash.Sum(nil))
	if !strings.EqualFold(actualChecksum, expectedChecksum) {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
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
