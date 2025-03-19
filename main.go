package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/noamstrauss/ota-updater/config"
	"github.com/noamstrauss/ota-updater/updater"
	"github.com/noamstrauss/ota-updater/version"
)

var (
	configPath = flag.String("config", "./config.json", "Path to config file")
)

func main() {
	flag.Parse()
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ldate | log.Ltime)
	log.Printf("Starting application version %s", version.Version)

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// Handle termination signals (Ctrl+C)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	// Run updater and application
	go runUpdateChecker(ctx, cfg)
	go runApplication(ctx)

	// Wait for termination signal
	<-sigs
	log.Println("Shutdown signal received, exiting...")

	// Cancel context to stop goroutines
	cancel()

	// Give time for cleanup before exit
	time.Sleep(2 * time.Second)
	log.Println("Application exited")
}

// runUpdateChecker periodically checks for updates
func runUpdateChecker(ctx context.Context, cfg *config.Config) {
	ticker := time.NewTicker(cfg.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping update checker...")
			return
		case <-ticker.C:
			log.Println("Checking for updates...")
			updateConfig := updater.Config{
				CurrentVersion: version.Version,
				GithubRepo:     cfg.GithubRepo,
				GithubToken:    cfg.GithubToken,
				ExecutablePath: os.Args[0],
			}

			hasUpdate, err := updater.CheckAndUpdate(updateConfig)
			if err != nil {
				log.Printf("Update error: %v", err)
			} else if hasUpdate {
				log.Println("Application updated successfully. Restarting...")
				updater.RestartApplication(os.Args[0], os.Args[1:])
			} else {
				log.Println("No updates available")
			}
		}
	}
}

func runApplication(ctx context.Context) {
	log.Println("Application is running...")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping application loop...")
			return
		case <-ticker.C:
			log.Println("Application still running...")
		}
	}
}
