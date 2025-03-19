// main.go
package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/noamstrauss/ota-updater/config"
	"github.com/noamstrauss/ota-updater/updater"
	"github.com/noamstrauss/ota-updater/version"
)

var (
	configPath = flag.String("config", "./config.json", "Path to configuration file")
	noUpdate   = flag.Bool("no-update", false, "Disable auto-updates")
)

func main() {
	flag.Parse()

	// Configure logging
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime)
	log.Printf("Starting application version %s", version.Version)

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override update setting if --no-update flag is set
	if *noUpdate {
		cfg.UpdateEnabled = false
	}

	// Start the updater in the background if enabled
	if cfg.UpdateEnabled {
		go func() {
			for {
				checkForUpdates(cfg)
				time.Sleep(cfg.UpdateInterval)
			}
		}()
	} else {
		log.Println("Auto-updates are disabled")
	}

	// Run your actual application here
	runApplication(cfg)
}

func checkForUpdates(cfg *config.Config) {
	log.Println("Checking for updates...")

	// Create update configuration
	updateConfig := updater.Config{
		CurrentVersion:  version.Version,
		GithubRepo:      cfg.GithubRepo,
		GithubToken:     cfg.GithubToken,
		CheckPrerelease: cfg.CheckPrerelease,
		ExecutablePath:  os.Args[0],
	}

	// Check for updates
	hasUpdate, err := updater.CheckAndUpdate(updateConfig)
	if err != nil {
		log.Printf("Update error: %v", err)
		return
	}

	if hasUpdate {
		log.Println("Application updated successfully. Restarting...")
		updater.RestartApplication(os.Args[0], os.Args[1:])
	} else {
		log.Println("No updates available")
	}
}

func runApplication(cfg *config.Config) {
	log.Println("Application is running...")

	// Ensure data directory exists
	if err := os.MkdirAll(cfg.DataDir, 0755); err != nil {
		log.Printf("Failed to create data directory: %v", err)
	}

	// Your application logic here
	// This is just a placeholder to keep the app running
	select {}
}
