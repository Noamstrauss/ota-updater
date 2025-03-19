package main

import (
	"flag"
	"github.com/noamstrauss/ota-updater/config"
	"github.com/noamstrauss/ota-updater/updater"
	"github.com/noamstrauss/ota-updater/version"
	"log"
	"os"
	"time"
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

	// Create a done channel to signal when the application should exit
	done := make(chan bool)

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
	go runApplication(cfg)

	// Wait for termination signal
	<-done
}

func checkForUpdates(cfg *config.Config) {
	log.Println("Checking for updates...")

	// Create update configuration
	updateConfig := updater.Config{
		CurrentVersion: version.Version,
		GithubRepo:     cfg.GithubRepo,
		GithubToken:    cfg.GithubToken,
		ExecutablePath: os.Args[0],
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

	// Simulate actual application work
	for {
		// Here you would put your application's main loop logic
		time.Sleep(5 * time.Second)
		log.Println("Application still running...")
	}
}
