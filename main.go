package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github-monitor/api"
	"github-monitor/config"
	"github-monitor/db"
	"github-monitor/github"
	"github-monitor/monitor"
)

func main() {
	// Load configuration
	if err := config.LoadConfig("config.yaml"); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	if err := db.InitDB(&config.AppConfig.Database); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Run migrations
	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize GitHub token pool with proxy config
	proxyConfig := &github.ProxyConfig{
		Enabled:  config.AppConfig.GitHub.ProxyEnabled,
		URL:      config.AppConfig.GitHub.ProxyURL,
		Type:     config.AppConfig.GitHub.ProxyType,
		Username: config.AppConfig.GitHub.ProxyUsername,
		Password: config.AppConfig.GitHub.ProxyPassword,
	}
	tokenPool, err := github.NewTokenPool(config.AppConfig.GitHub.Tokens, proxyConfig)
	if err != nil {
		log.Fatalf("Failed to initialize token pool: %v", err)
	}

	// Refresh token information
	ctx := context.Background()
	tokenPool.RefreshAllTokens(ctx)

	// Initialize search service
	searchService := github.NewSearchService(tokenPool)

	// Parse scan interval
	scanInterval, err := time.ParseDuration(config.AppConfig.Monitor.ScanInterval)
	if err != nil {
		log.Printf("Invalid scan interval, using default 5 minutes: %v", err)
		scanInterval = 5 * time.Minute
	}

	// Initialize monitor service
	monitorService := monitor.NewMonitorService(searchService, scanInterval)

	// Start monitor if enabled
	if config.AppConfig.Monitor.Enabled {
		monitorService.Start()
	}

	// Initialize API
	apiService := api.NewAPI(tokenPool, searchService, monitorService)
	router := api.SetupRouter(apiService)

	// Start server
	addr := fmt.Sprintf(":%d", config.AppConfig.Server.Port)
	log.Printf("Starting server on %s", addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
