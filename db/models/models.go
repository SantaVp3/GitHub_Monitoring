package models

import (
	"time"

	"gorm.io/gorm"
)

// GitHubToken represents a GitHub API token
type GitHubToken struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	Token        string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"token"`
	Name         string         `gorm:"type:varchar(255)" json:"name"`
	RateLimit    int            `json:"rate_limit"`
	RateRemaining int           `json:"rate_remaining"`
	RateReset    time.Time      `json:"rate_reset"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	LastUsed     *time.Time     `json:"last_used"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// MonitorRule represents a monitoring rule with keywords
type MonitorRule struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	Keywords    string         `gorm:"type:text;not null" json:"keywords"` // JSON array of keywords
	MatchType   string         `gorm:"type:varchar(50);default:'fuzzy'" json:"match_type"` // "precise" or "fuzzy"
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	ExcludeExts string         `gorm:"type:text" json:"exclude_exts"` // JSON array of file extensions to exclude
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// SearchResult represents a search result from GitHub
type SearchResult struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	RuleID       uint           `gorm:"index;not null" json:"rule_id"`
	Rule         MonitorRule    `gorm:"foreignKey:RuleID" json:"rule,omitempty"`
	RepoFullName string         `gorm:"type:varchar(255);index;not null" json:"repo_full_name"`
	RepoURL      string         `gorm:"type:varchar(512)" json:"repo_url"`
	FilePath     string         `gorm:"type:varchar(512)" json:"file_path"`
	FileURL      string         `gorm:"type:varchar(512)" json:"file_url"`
	MatchedKeywords string      `gorm:"type:text" json:"matched_keywords"` // JSON array
	ContentSnippet  string      `gorm:"type:text" json:"content_snippet"`
	HTMLURL      string         `gorm:"type:varchar(512)" json:"html_url"`
	Score        float64        `json:"score"`
	Status       string         `gorm:"type:varchar(50);default:'pending'" json:"status"` // pending, reviewed, false_positive, confirmed
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// Whitelist represents whitelisted repositories or users
type Whitelist struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Type        string         `gorm:"type:varchar(50);not null" json:"type"` // "user" or "repo"
	Value       string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"value"`
	Description string         `gorm:"type:text" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// ScanHistory represents monitoring scan history
type ScanHistory struct {
	ID           uint      `gorm:"primarykey" json:"id"`
	RuleID       uint      `gorm:"index;not null" json:"rule_id"`
	Rule         MonitorRule `gorm:"foreignKey:RuleID" json:"rule,omitempty"`
	ResultsCount int       `json:"results_count"`
	NewResults   int       `json:"new_results"`
	TokenUsed    string    `gorm:"type:varchar(100)" json:"token_used"`
	Status       string    `gorm:"type:varchar(50);default:'success'" json:"status"` // success, failed, rate_limited
	ErrorMessage string    `gorm:"type:text" json:"error_message"`
	Duration     int       `json:"duration"` // in seconds
	CreatedAt    time.Time `json:"created_at"`
}

// NotificationConfig represents notification settings
type NotificationConfig struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Type        string         `gorm:"type:varchar(50);not null" json:"type"` // wecom, dingtalk, feishu, webhook
	Enabled     bool           `gorm:"default:false" json:"enabled"`
	WebhookURL  string         `gorm:"type:varchar(512)" json:"webhook_url"`
	Secret      string         `gorm:"type:varchar(255)" json:"secret,omitempty"`
	NotifyOnNew bool           `gorm:"default:true" json:"notify_on_new"`     // Notify on new leaks
	NotifyOnConfirmed bool    `gorm:"default:true" json:"notify_on_confirmed"` // Notify on confirmed leaks
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
