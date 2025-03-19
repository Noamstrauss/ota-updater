// main.go

package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/noamstrauss/ota-updater/updater"
	"github.com/noamstrauss/ota-updater/version"
)

var (
	updateInterval = flag.Duration("update-interval", 1*time.Minute, "Interval between update checks")
	updateURL      = flag.String("update-url", "http://localhost:8080", "URL for update server")
)

func main() {
	flag.Parse()

	// Configure logging
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime)
	log.Printf("Starting application version %s", version.Version)

	// Start the updater in the background
	go func() {
		for {
			checkForUpdates()
			time.Sleep(*updateInterval)
		}
	}()

	// Run your actual application here
	runApplication()
}

func checkForUpdates() {
	log.Println("Checking for updates...")

	// Create update configuration
	config := updater.Config{
		CurrentVersion: version.Version,
		UpdateURL:      *updateURL,
		ExecutablePath: os.Args[0],
	}

	// Check for updates
	hasUpdate, err := updater.CheckAndUpdate(config)
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

func runApplication() {
	log.Println("Application is running...")

	// Your application logic here
	// This is just a placeholder to keep the app running
	select {}
}
