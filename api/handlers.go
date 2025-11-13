package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github-monitor/auth"
	"github-monitor/db"
	"github-monitor/db/models"
	"github-monitor/github"
	"github-monitor/monitor"

	"github.com/gin-gonic/gin"
)

type API struct {
	tokenPool      *github.TokenPool
	searchService  *github.SearchService
	monitorService *monitor.MonitorService
}

func NewAPI(tokenPool *github.TokenPool, searchService *github.SearchService, monitorService *monitor.MonitorService) *API {
	return &API{
		tokenPool:      tokenPool,
		searchService:  searchService,
		monitorService: monitorService,
	}
}

// GetTokens returns all GitHub tokens
func (a *API) GetTokens(c *gin.Context) {
	var tokens []models.GitHubToken
	if err := db.GetDB().Find(&tokens).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tokens)
}

// CreateToken creates a new GitHub token
func (a *API) CreateToken(c *gin.Context) {
	var token models.GitHubToken
	if err := c.ShouldBindJSON(&token); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.GetDB().Create(&token).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, token)
}

// DeleteToken deletes a token
func (a *API) DeleteToken(c *gin.Context) {
	id := c.Param("id")
	if err := db.GetDB().Delete(&models.GitHubToken{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Token deleted successfully"})
}

// GetTokenStats returns statistics about all tokens in the pool
func (a *API) GetTokenStats(c *gin.Context) {
	stats := a.tokenPool.GetTokenStats()
	c.JSON(http.StatusOK, stats)
}

// GetMonitorRules returns all monitor rules
func (a *API) GetMonitorRules(c *gin.Context) {
	var rules []models.MonitorRule
	if err := db.GetDB().Find(&rules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rules)
}

// GetMonitorRule returns a single monitor rule
func (a *API) GetMonitorRule(c *gin.Context) {
	id := c.Param("id")
	var rule models.MonitorRule
	if err := db.GetDB().First(&rule, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// CreateMonitorRule creates a new monitor rule
func (a *API) CreateMonitorRule(c *gin.Context) {
	var rule models.MonitorRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate keywords JSON
	if rule.Keywords != "" {
		var keywords []string
		if err := json.Unmarshal([]byte(rule.Keywords), &keywords); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid keywords JSON format"})
			return
		}
	}

	if err := db.GetDB().Create(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

// UpdateMonitorRule updates a monitor rule
func (a *API) UpdateMonitorRule(c *gin.Context) {
	id := c.Param("id")
	var rule models.MonitorRule

	if err := db.GetDB().First(&rule, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
		return
	}

	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.GetDB().Save(&rule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// DeleteMonitorRule deletes a monitor rule
func (a *API) DeleteMonitorRule(c *gin.Context) {
	id := c.Param("id")
	if err := db.GetDB().Delete(&models.MonitorRule{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rule deleted successfully"})
}

// GetSearchResults returns search results with pagination
func (a *API) GetSearchResults(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	ruleID := c.Query("rule_id")
	status := c.Query("status")

	offset := (page - 1) * pageSize

	query := db.GetDB().Model(&models.SearchResult{})

	if ruleID != "" {
		query = query.Where("rule_id = ?", ruleID)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var results []models.SearchResult
	if err := query.Preload("Rule").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"results":   results,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// UpdateSearchResult updates a search result status
func (a *API) UpdateSearchResult(c *gin.Context) {
	id := c.Param("id")
	var result models.SearchResult

	if err := db.GetDB().First(&result, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Result not found"})
		return
	}

	var input struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result.Status = input.Status

	if err := db.GetDB().Save(&result).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// BatchUpdateSearchResults updates multiple search results at once
func (a *API) BatchUpdateSearchResults(c *gin.Context) {
	var input struct {
		IDs    []uint `json:"ids" binding:"required"`
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(input.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No IDs provided"})
		return
	}

	// Validate status
	validStatuses := map[string]bool{
		"pending":        true,
		"confirmed":      true,
		"false_positive": true,
	}

	if !validStatuses[input.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return
	}

	// Update all results
	if err := db.GetDB().Model(&models.SearchResult{}).
		Where("id IN ?", input.IDs).
		Update("status", input.Status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Batch update successful",
		"updated": len(input.IDs),
	})
}

// GetWhitelist returns all whitelist entries
func (a *API) GetWhitelist(c *gin.Context) {
	var whitelist []models.Whitelist
	if err := db.GetDB().Find(&whitelist).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, whitelist)
}

// CreateWhitelist creates a new whitelist entry
func (a *API) CreateWhitelist(c *gin.Context) {
	var entry models.Whitelist
	if err := c.ShouldBindJSON(&entry); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.GetDB().Create(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, entry)
}

// DeleteWhitelist deletes a whitelist entry
func (a *API) DeleteWhitelist(c *gin.Context) {
	id := c.Param("id")
	if err := db.GetDB().Delete(&models.Whitelist{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Whitelist entry deleted successfully"})
}

// GetScanHistory returns scan history
func (a *API) GetScanHistory(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	ruleID := c.Query("rule_id")

	offset := (page - 1) * pageSize

	query := db.GetDB().Model(&models.ScanHistory{})

	if ruleID != "" {
		query = query.Where("rule_id = ?", ruleID)
	}

	var total int64
	query.Count(&total)

	var history []models.ScanHistory
	if err := query.Preload("Rule").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&history).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history":   history,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetMonitorStatus returns monitor service status
func (a *API) GetMonitorStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"is_running": a.monitorService.IsRunning(),
	})
}

// StartMonitor starts the monitoring service
func (a *API) StartMonitor(c *gin.Context) {
	if a.monitorService.IsRunning() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Monitor is already running"})
		return
	}

	a.monitorService.Start()
	c.JSON(http.StatusOK, gin.H{"message": "Monitor started successfully"})
}

// StopMonitor stops the monitoring service
func (a *API) StopMonitor(c *gin.Context) {
	if !a.monitorService.IsRunning() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Monitor is not running"})
		return
	}

	a.monitorService.Stop()
	c.JSON(http.StatusOK, gin.H{"message": "Monitor stopped successfully"})
}

// GetDashboardStats returns dashboard statistics
func (a *API) GetDashboardStats(c *gin.Context) {
	var stats struct {
		TotalRules       int64 `json:"total_rules"`
		ActiveRules      int64 `json:"active_rules"`
		TotalResults     int64 `json:"total_results"`
		PendingResults   int64 `json:"pending_results"`
		ConfirmedResults int64 `json:"confirmed_results"`
		TotalTokens      int64 `json:"total_tokens"`
		ActiveTokens     int64 `json:"active_tokens"`
	}

	db.GetDB().Model(&models.MonitorRule{}).Count(&stats.TotalRules)
	db.GetDB().Model(&models.MonitorRule{}).Where("is_active = ?", true).Count(&stats.ActiveRules)
	db.GetDB().Model(&models.SearchResult{}).Count(&stats.TotalResults)
	db.GetDB().Model(&models.SearchResult{}).Where("status = ?", "pending").Count(&stats.PendingResults)
	db.GetDB().Model(&models.SearchResult{}).Where("status = ?", "confirmed").Count(&stats.ConfirmedResults)
	db.GetDB().Model(&models.GitHubToken{}).Count(&stats.TotalTokens)
	db.GetDB().Model(&models.GitHubToken{}).Where("is_active = ?", true).Count(&stats.ActiveTokens)

	c.JSON(http.StatusOK, stats)
}

// Notification handlers

// GetNotifications returns all notification configs
func (a *API) GetNotifications(c *gin.Context) {
	var notifications []models.NotificationConfig
	if err := db.GetDB().Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notifications)
}

// CreateNotification creates a new notification config
func (a *API) CreateNotification(c *gin.Context) {
	var notification models.NotificationConfig
	if err := c.ShouldBindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.GetDB().Create(&notification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, notification)
}

// UpdateNotification updates a notification config
func (a *API) UpdateNotification(c *gin.Context) {
	id := c.Param("id")
	var notification models.NotificationConfig

	if err := db.GetDB().First(&notification, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	if err := c.ShouldBindJSON(&notification); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.GetDB().Save(&notification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notification)
}

// DeleteNotification deletes a notification config
func (a *API) DeleteNotification(c *gin.Context) {
	id := c.Param("id")
	if err := db.GetDB().Delete(&models.NotificationConfig{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification deleted successfully"})
}

// TestNotification sends a test notification
func (a *API) TestNotification(c *gin.Context) {
	id := c.Param("id")
	var notification models.NotificationConfig

	if err := db.GetDB().First(&notification, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	// Import notify package here - will be added in router
	c.JSON(http.StatusOK, gin.H{"message": "Test notification functionality - implement in router"})
}

// Login handles user login
func (a *API) Login(c *gin.Context) {
	var input struct {
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}

	// Verify password
	if !auth.VerifyPassword(input.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}

	// Generate token
	token, err := auth.GenerateToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"message": "Login successful",
	})
}

// GetAuthStatus returns the current authentication status
func (a *API) GetAuthStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
	})
}
