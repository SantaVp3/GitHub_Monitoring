package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	GitHub   GitHubConfig   `mapstructure:"github"`
	Monitor  MonitorConfig  `mapstructure:"monitor"`
	Auth     AuthConfig     `mapstructure:"auth"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

type GitHubConfig struct {
	Tokens              []string `mapstructure:"tokens"`
	RateLimitThreshold  int      `mapstructure:"rate_limit_threshold"`
	RequestInterval     string   `mapstructure:"request_interval"`
	ProxyEnabled        bool     `mapstructure:"proxy_enabled"`
	ProxyURL            string   `mapstructure:"proxy_url"`
	ProxyType           string   `mapstructure:"proxy_type"` // http, https, socks5
	ProxyUsername       string   `mapstructure:"proxy_username"`
	ProxyPassword       string   `mapstructure:"proxy_password"`
}

type MonitorConfig struct {
	Enabled      bool   `mapstructure:"enabled"`
	ScanInterval string `mapstructure:"scan_interval"`
}

type AuthConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Password   string `mapstructure:"password"`
	JWTSecret  string `mapstructure:"jwt_secret"`
	TokenExpiry string `mapstructure:"token_expiry"` // e.g., "24h", "7d"
}

var AppConfig *Config

func LoadConfig(configPath string) error {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	viper.SetDefault("server.port", 8080)
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("github.rate_limit_threshold", 10)
	viper.SetDefault("github.request_interval", "5s")
	viper.SetDefault("monitor.enabled", true)
	viper.SetDefault("monitor.scan_interval", "300s")
	viper.SetDefault("auth.enabled", false)
	viper.SetDefault("auth.token_expiry", "24h")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	AppConfig = &Config{}
	if err := viper.Unmarshal(AppConfig); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	log.Println("Configuration loaded successfully")
	return nil
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
	)
}
