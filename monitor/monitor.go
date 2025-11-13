package monitor

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github-monitor/db"
	"github-monitor/db/models"
	"github-monitor/github"
)

// MonitorService handles the monitoring logic
type MonitorService struct {
	searchService *github.SearchService
	scanInterval  time.Duration
	isRunning     bool
	stopChan      chan bool
}

// NewMonitorService creates a new monitor service
func NewMonitorService(searchService *github.SearchService, scanInterval time.Duration) *MonitorService {
	return &MonitorService{
		searchService: searchService,
		scanInterval:  scanInterval,
		isRunning:     false,
		stopChan:      make(chan bool),
	}
}

// Start starts the monitoring service
func (m *MonitorService) Start() {
	if m.isRunning {
		log.Println("Monitor service is already running")
		return
	}

	m.isRunning = true
	log.Println("Monitor service started")

	go m.run()
}

// Stop stops the monitoring service
func (m *MonitorService) Stop() {
	if !m.isRunning {
		return
	}

	log.Println("Stopping monitor service...")
	m.stopChan <- true
	m.isRunning = false
	log.Println("Monitor service stopped")
}

// IsRunning returns whether the monitor is running
func (m *MonitorService) IsRunning() bool {
	return m.isRunning
}

// run is the main monitoring loop
func (m *MonitorService) run() {
	ticker := time.NewTicker(m.scanInterval)
	defer ticker.Stop()

	// Run initial scan
	m.scan()

	for {
		select {
		case <-ticker.C:
			m.scan()
		case <-m.stopChan:
			return
		}
	}
}

// scan performs a single scan of all active rules
func (m *MonitorService) scan() {
	log.Println("Starting monitoring scan...")
	ctx := context.Background()

	// Get all active rules
	var rules []models.MonitorRule
	if err := db.GetDB().Where("is_active = ?", true).Find(&rules).Error; err != nil {
		log.Printf("Failed to fetch monitor rules: %v", err)
		return
	}

	log.Printf("Found %d active monitoring rules", len(rules))

	for _, rule := range rules {
		m.scanRule(ctx, rule)
		// Wait between rules to avoid overwhelming the API
		time.Sleep(5 * time.Second)
	}

	log.Println("Monitoring scan completed")
}

// scanRule scans a single monitoring rule
func (m *MonitorService) scanRule(ctx context.Context, rule models.MonitorRule) {
	startTime := time.Now()
	log.Printf("Scanning rule: %s (ID: %d)", rule.Name, rule.ID)

	// Parse keywords
	keywords, err := github.ParseKeywords(rule.Keywords)
	if err != nil {
		log.Printf("Failed to parse keywords for rule %d: %v", rule.ID, err)
		m.recordScanHistory(rule.ID, 0, 0, "", "failed", err.Error(), 0)
		return
	}

	// Parse exclude extensions
	excludeExts, err := github.ParseExcludeExts(rule.ExcludeExts)
	if err != nil {
		log.Printf("Failed to parse exclude extensions for rule %d: %v", rule.ID, err)
		excludeExts = []string{}
	}

	// Build search options
	searchOpts := github.SearchOptions{
		Keywords:    keywords,
		MatchType:   rule.MatchType,
		ExcludeExts: excludeExts,
		Sort:        "indexed",
		Order:       "desc",
	}

	// Perform search
	results, err := m.searchService.SearchWithRetry(ctx, searchOpts, 3)
	if err != nil {
		log.Printf("Search failed for rule %d: %v", rule.ID, err)
		status := "failed"
		if err.Error() == "rate limit exceeded" {
			status = "rate_limited"
		}
		duration := int(time.Since(startTime).Seconds())
		m.recordScanHistory(rule.ID, 0, 0, "", status, err.Error(), duration)
		return
	}

	// Filter results against whitelist
	filteredResults := m.filterWhitelist(results)

	// Save new results
	newResultsCount := m.saveResults(rule.ID, filteredResults)

	duration := int(time.Since(startTime).Seconds())
	log.Printf("Rule %d scan completed: %d results found, %d new results, took %d seconds",
		rule.ID, len(filteredResults), newResultsCount, duration)

	m.recordScanHistory(rule.ID, len(filteredResults), newResultsCount, "", "success", "", duration)
}

// filterWhitelist filters results against the whitelist
func (m *MonitorService) filterWhitelist(results []*github.SearchResultItem) []*github.SearchResultItem {
	var whitelist []models.Whitelist
	if err := db.GetDB().Find(&whitelist).Error; err != nil {
		log.Printf("Failed to fetch whitelist: %v", err)
		return results
	}

	if len(whitelist) == 0 {
		return results
	}

	filtered := make([]*github.SearchResultItem, 0)

	for _, result := range results {
		isWhitelisted := false

		for _, entry := range whitelist {
			if entry.Type == "repo" && result.RepoFullName == entry.Value {
				isWhitelisted = true
				break
			}
			// For user type, check if repo belongs to that user
			if entry.Type == "user" {
				// RepoFullName format: "username/reponame"
				if len(result.RepoFullName) > 0 {
					parts := splitRepoName(result.RepoFullName)
					if len(parts) > 0 && parts[0] == entry.Value {
						isWhitelisted = true
						break
					}
				}
			}
		}

		if !isWhitelisted {
			filtered = append(filtered, result)
		}
	}

	log.Printf("Whitelist filtering: %d -> %d results", len(results), len(filtered))
	return filtered
}

// splitRepoName splits a full repo name into user and repo parts
func splitRepoName(fullName string) []string {
	parts := make([]string, 0)
	for i := 0; i < len(fullName); i++ {
		if fullName[i] == '/' {
			parts = append(parts, fullName[:i])
			parts = append(parts, fullName[i+1:])
			break
		}
	}
	return parts
}

// saveResults saves search results to database
func (m *MonitorService) saveResults(ruleID uint, results []*github.SearchResultItem) int {
	newCount := 0

	for _, result := range results {
		// Check if result already exists
		var existingResult models.SearchResult
		err := db.GetDB().Where("rule_id = ? AND repo_full_name = ? AND file_path = ?",
			ruleID, result.RepoFullName, result.FilePath).First(&existingResult).Error

		if err != nil {
			// Result doesn't exist, create new one
			matchedKeywordsJSON, _ := json.Marshal(result.MatchedKeywords)

			newResult := models.SearchResult{
				RuleID:          ruleID,
				RepoFullName:    result.RepoFullName,
				RepoURL:         result.RepoURL,
				FilePath:        result.FilePath,
				FileURL:         result.FileURL,
				MatchedKeywords: string(matchedKeywordsJSON),
				ContentSnippet:  result.ContentSnippet,
				HTMLURL:         result.HTMLURL,
				Score:           result.Score,
				Status:          "pending",
			}

			if err := db.GetDB().Create(&newResult).Error; err != nil {
				log.Printf("Failed to save result: %v", err)
			} else {
				newCount++
			}
		}
	}

	return newCount
}

// recordScanHistory records a scan history entry
func (m *MonitorService) recordScanHistory(ruleID uint, resultsCount, newResults int, tokenUsed, status, errorMsg string, duration int) {
	history := models.ScanHistory{
		RuleID:       ruleID,
		ResultsCount: resultsCount,
		NewResults:   newResults,
		TokenUsed:    tokenUsed,
		Status:       status,
		ErrorMessage: errorMsg,
		Duration:     duration,
	}

	if err := db.GetDB().Create(&history).Error; err != nil {
		log.Printf("Failed to record scan history: %v", err)
	}
}
