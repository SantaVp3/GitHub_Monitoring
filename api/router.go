package api

import (
	"net/http"
	"path/filepath"

	"github-monitor/auth"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(api *API) *gin.Engine {
	r := gin.Default()

	// CORS middleware
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000", "http://localhost:5173"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Public routes (no authentication required)
	public := r.Group("/api/v1")
	{
		public.POST("/login", api.Login)
	}

	// Protected API routes (require authentication)
	v1 := r.Group("/api/v1")
	v1.Use(auth.AuthMiddleware())
	{
		// Auth
		v1.GET("/auth/status", api.GetAuthStatus)

		// Dashboard
		v1.GET("/dashboard/stats", api.GetDashboardStats)

		// Tokens
		tokens := v1.Group("/tokens")
		{
			tokens.GET("", api.GetTokens)
			tokens.POST("", api.CreateToken)
			tokens.DELETE("/:id", api.DeleteToken)
			tokens.GET("/stats", api.GetTokenStats)
		}

		// Monitor rules
		rules := v1.Group("/rules")
		{
			rules.GET("", api.GetMonitorRules)
			rules.GET("/:id", api.GetMonitorRule)
			rules.POST("", api.CreateMonitorRule)
			rules.PUT("/:id", api.UpdateMonitorRule)
			rules.DELETE("/:id", api.DeleteMonitorRule)
		}

		// Search results
		results := v1.Group("/results")
		{
			results.GET("", api.GetSearchResults)
			results.PUT("/:id", api.UpdateSearchResult)
			results.POST("/batch", api.BatchUpdateSearchResults)
		}

		// Whitelist
		whitelist := v1.Group("/whitelist")
		{
			whitelist.GET("", api.GetWhitelist)
			whitelist.POST("", api.CreateWhitelist)
			whitelist.DELETE("/:id", api.DeleteWhitelist)
		}

		// Scan history
		v1.GET("/history", api.GetScanHistory)

		// Monitor control
		monitor := v1.Group("/monitor")
		{
			monitor.GET("/status", api.GetMonitorStatus)
			monitor.POST("/start", api.StartMonitor)
			monitor.POST("/stop", api.StopMonitor)
		}

		// Notifications
		notifications := v1.Group("/notifications")
		{
			notifications.GET("", api.GetNotifications)
			notifications.POST("", api.CreateNotification)
			notifications.PUT("/:id", api.UpdateNotification)
			notifications.DELETE("/:id", api.DeleteNotification)
			notifications.POST("/:id/test", api.TestNotification)
		}
	}

	// Serve static files from frontend dist
	distPath := filepath.Join(".", "frontend", "dist")
	r.Static("/assets", filepath.Join(distPath, "assets"))

	// Serve index.html for all non-API routes (SPA catch-all)
	r.NoRoute(func(c *gin.Context) {
		// Don't serve index.html for API routes
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}
		c.File(filepath.Join(distPath, "index.html"))
	})

	return r
}
