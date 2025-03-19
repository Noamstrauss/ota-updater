// server/server.go

package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	port        = flag.Int("port", 8080, "Server port")
	releasesDir = flag.String("releases-dir", "./releases", "Directory containing releases")
)

// VersionInfo represents metadata about a release
type VersionInfo struct {
	Version     string `json:"version"`
	ReleaseDate string `json:"release_date"`
	DownloadURL string `json:"download_url"`
	Checksum    string `json:"checksum"`
}

func main() {
	flag.Parse()

	// Create releases directory if it doesn't exist
	if err := os.MkdirAll(*releasesDir, 0755); err != nil {
		log.Fatalf("Failed to create releases directory: %v", err)
	}

	// Define handlers
	http.HandleFunc("/check/", checkHandler)
	http.HandleFunc("/download/", downloadHandler)
	http.HandleFunc("/upload", uploadHandler)

	// Start server
	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting update server on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

// checkHandler checks if an update is available
func checkHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid request path", http.StatusBadRequest)
		return
	}

	platform := parts[2]
	arch := parts[3]
	currentVersion := parts[4]

	// Find latest version
	latestVersion, versionInfo, err := findLatestVersion(platform, arch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// No update available
	if latestVersion == "" || latestVersion <= currentVersion {
		http.NotFound(w, r)
		return
	}

	// Return version info
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(versionInfo)
}

// downloadHandler serves binary downloads
func downloadHandler(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "Invalid request path", http.StatusBadRequest)
		return
	}

	platform := parts[2]
	arch := parts[3]
	version := parts[4]

	// Build file path
	filePath := filepath.Join(*releasesDir, platform, arch, version+".bin")

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	// Serve the file
	http.ServeFile(w, r, filePath)
}

// uploadHandler handles uploading a new release
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get form values
	platform := r.FormValue("platform")
	arch := r.FormValue("arch")
	version := r.FormValue("version")

	if platform == "" || arch == "" || version == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Get uploaded file
	file, _, err := r.FormFile("binary")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create directory structure
	dirPath := filepath.Join(*releasesDir, platform, arch)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Save binary file
	binPath := filepath.Join(dirPath, version+".bin")
	dst, err := os.Create(binPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate checksum while copying
	hash := sha256.New()
	multiWriter := io.MultiWriter(dst, hash)

	if _, err := io.Copy(multiWriter, file); err != nil {
		dst.Close()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	dst.Close()

	// Set executable permissions
	if err := os.Chmod(binPath, 0755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate checksum
	checksum := hex.EncodeToString(hash.Sum(nil))

	// Create version info
	versionInfo := VersionInfo{
		Version:     version,
		ReleaseDate: time.Now().Format(time.RFC3339),
		DownloadURL: fmt.Sprintf("/download/%s/%s/%s", platform, arch, version),
		Checksum:    checksum,
	}

	// Save metadata
	metaPath := filepath.Join(dirPath, version+".json")
	metaFile, err := os.Create(metaPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer metaFile.Close()

	if err := json.NewEncoder(metaFile).Encode(versionInfo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return success
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "Version %s uploaded successfully", version)
}

// findLatestVersion finds the latest version for a platform/arch
func findLatestVersion(platform, arch string) (string, VersionInfo, error) {
	dirPath := filepath.Join(*releasesDir, platform, arch)
	files, err := os.ReadDir(dirPath)
	if os.IsNotExist(err) {
		return "", VersionInfo{}, nil
	}
	if err != nil {
		return "", VersionInfo{}, err
	}

	var latestVersion string
	var versionInfo VersionInfo

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		// Extract version from filename
		version := strings.TrimSuffix(file.Name(), ".json")

		// Read version info
		data, err := os.ReadFile(filepath.Join(dirPath, file.Name()))
		if err != nil {
			continue
		}

		var info VersionInfo
		if err := json.Unmarshal(data, &info); err != nil {
			continue
		}

		// Check if this is the latest version
		if latestVersion == "" || version > latestVersion {
			latestVersion = version
			versionInfo = info
		}
	}

	return latestVersion, versionInfo, nil
}
