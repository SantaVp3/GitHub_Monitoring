package github

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/google/go-github/v57/github"
	"golang.org/x/net/proxy"
	"golang.org/x/oauth2"
)

// ProxyConfig holds proxy configuration
type ProxyConfig struct {
	Enabled  bool
	URL      string
	Type     string // http, https, socks5
	Username string
	Password string
}

// TokenPool manages multiple GitHub tokens with automatic rotation
type TokenPool struct {
	tokens       []*TokenInfo
	currentIndex int
	proxyConfig  *ProxyConfig
	mu           sync.RWMutex
}

// TokenInfo holds information about a GitHub token
type TokenInfo struct {
	Token       string
	Client      *github.Client
	RateLimit   *github.Rate
	IsAvailable bool
	LastChecked time.Time
	mu          sync.RWMutex
}

// NewTokenPool creates a new token pool
func NewTokenPool(tokens []string, proxyConfig *ProxyConfig) (*TokenPool, error) {
	if len(tokens) == 0 {
		return nil, fmt.Errorf("no tokens provided")
	}

	pool := &TokenPool{
		tokens:       make([]*TokenInfo, 0, len(tokens)),
		currentIndex: 0,
		proxyConfig:  proxyConfig,
	}

	for _, token := range tokens {
		if token == "" {
			continue
		}

		tokenInfo := &TokenInfo{
			Token:       token,
			Client:      createClient(token, proxyConfig),
			IsAvailable: true,
			LastChecked: time.Now(),
		}

		pool.tokens = append(pool.tokens, tokenInfo)
	}

	if len(pool.tokens) == 0 {
		return nil, fmt.Errorf("no valid tokens provided")
	}

	log.Printf("Token pool initialized with %d tokens", len(pool.tokens))
	if proxyConfig != nil && proxyConfig.Enabled {
		log.Printf("Proxy enabled: %s (%s)", proxyConfig.URL, proxyConfig.Type)
	}
	return pool, nil
}

// createClient creates a GitHub client with the given token and proxy config
func createClient(token string, proxyConfig *ProxyConfig) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)

	// Create HTTP transport
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
	}

	// Configure proxy if enabled
	if proxyConfig != nil && proxyConfig.Enabled && proxyConfig.URL != "" {
		if proxyConfig.Type == "socks5" {
			// SOCKS5 proxy
			proxyURL, err := url.Parse(proxyConfig.URL)
			if err == nil {
				var auth *proxy.Auth
				if proxyConfig.Username != "" {
					auth = &proxy.Auth{
						User:     proxyConfig.Username,
						Password: proxyConfig.Password,
					}
				}

				dialer, err := proxy.SOCKS5("tcp", proxyURL.Host, auth, proxy.Direct)
				if err == nil {
					transport.Dial = dialer.Dial
					log.Printf("SOCKS5 proxy configured: %s", proxyURL.Host)
				} else {
					log.Printf("Failed to configure SOCKS5 proxy: %v", err)
				}
			}
		} else {
			// HTTP/HTTPS proxy
			proxyURL, err := url.Parse(proxyConfig.URL)
			if err == nil {
				// Add auth if provided
				if proxyConfig.Username != "" {
					proxyURL.User = url.UserPassword(proxyConfig.Username, proxyConfig.Password)
				}
				transport.Proxy = http.ProxyURL(proxyURL)
				log.Printf("HTTP/HTTPS proxy configured: %s", proxyURL.Host)
			} else {
				log.Printf("Failed to parse proxy URL: %v", err)
			}
		}
	}

	// Create oauth2 client with custom HTTP transport
	tc := &http.Client{
		Transport: &oauth2.Transport{
			Source: ts,
			Base:   transport,
		},
	}

	return github.NewClient(tc)
}

// GetClient returns an available GitHub client
func (p *TokenPool) GetClient(ctx context.Context) (*github.Client, *TokenInfo, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	startIndex := p.currentIndex
	attempts := 0
	maxAttempts := len(p.tokens)

	for attempts < maxAttempts {
		tokenInfo := p.tokens[p.currentIndex]

		// Check if token is available
		if tokenInfo.IsAvailable {
			// Update rate limit info
			err := tokenInfo.UpdateRateLimit(ctx)
			if err != nil {
				log.Printf("Failed to update rate limit for token %d: %v", p.currentIndex, err)
				p.markTokenUnavailable(p.currentIndex)
				p.nextToken()
				attempts++
				continue
			}

			// Check if token has remaining calls
			if tokenInfo.HasRemainingCalls(10) { // Keep at least 10 calls in reserve
				log.Printf("Using token %d, remaining: %d/%d, resets at: %v",
					p.currentIndex,
					tokenInfo.RateLimit.Remaining,
					tokenInfo.RateLimit.Limit,
					tokenInfo.RateLimit.Reset.Time)
				return tokenInfo.Client, tokenInfo, nil
			}

			// Token is rate limited, mark as unavailable temporarily
			log.Printf("Token %d is rate limited, resets at: %v", p.currentIndex, tokenInfo.RateLimit.Reset.Time)
			p.markTokenUnavailable(p.currentIndex)
		}

		p.nextToken()
		attempts++

		// If we've cycled through all tokens, check if any will reset soon
		if p.currentIndex == startIndex && attempts == maxAttempts {
			nextReset := p.getNextResetTime()
			if !nextReset.IsZero() && time.Until(nextReset) < 5*time.Minute {
				log.Printf("All tokens exhausted, waiting until %v", nextReset)
				return nil, nil, fmt.Errorf("all tokens rate limited, next reset at %v", nextReset)
			}
		}
	}

	return nil, nil, fmt.Errorf("no available tokens")
}

// UpdateRateLimit updates the rate limit information for a token
func (t *TokenInfo) UpdateRateLimit(ctx context.Context) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	rateLimit, _, err := t.Client.RateLimit.Get(ctx)
	if err != nil {
		return err
	}

	if rateLimit != nil && rateLimit.Core != nil {
		t.RateLimit = rateLimit.Core
		t.LastChecked = time.Now()

		// Auto-recover if rate limit has reset
		if t.RateLimit.Remaining > 10 {
			t.IsAvailable = true
		}
	}

	return nil
}

// HasRemainingCalls checks if the token has enough remaining API calls
func (t *TokenInfo) HasRemainingCalls(threshold int) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.RateLimit == nil {
		return true
	}

	// Check if rate limit has reset
	if time.Now().After(t.RateLimit.Reset.Time) {
		return true
	}

	return t.RateLimit.Remaining > threshold
}

// markTokenUnavailable marks a token as unavailable
func (p *TokenPool) markTokenUnavailable(index int) {
	if index >= 0 && index < len(p.tokens) {
		p.tokens[index].mu.Lock()
		p.tokens[index].IsAvailable = false
		p.tokens[index].mu.Unlock()
	}
}

// nextToken moves to the next token in the pool
func (p *TokenPool) nextToken() {
	p.currentIndex = (p.currentIndex + 1) % len(p.tokens)
}

// getNextResetTime returns the earliest rate limit reset time
func (p *TokenPool) getNextResetTime() time.Time {
	var nextReset time.Time

	for _, tokenInfo := range p.tokens {
		tokenInfo.mu.RLock()
		if tokenInfo.RateLimit != nil {
			resetTime := tokenInfo.RateLimit.Reset.Time
			if nextReset.IsZero() || resetTime.Before(nextReset) {
				nextReset = resetTime
			}
		}
		tokenInfo.mu.RUnlock()
	}

	return nextReset
}

// GetTokenStats returns statistics about all tokens
func (p *TokenPool) GetTokenStats() []map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := make([]map[string]interface{}, len(p.tokens))

	for i, tokenInfo := range p.tokens {
		tokenInfo.mu.RLock()
		stat := map[string]interface{}{
			"index":       i,
			"is_available": tokenInfo.IsAvailable,
			"last_checked": tokenInfo.LastChecked,
		}

		if tokenInfo.RateLimit != nil {
			stat["rate_limit"] = tokenInfo.RateLimit.Limit
			stat["rate_remaining"] = tokenInfo.RateLimit.Remaining
			stat["rate_reset"] = tokenInfo.RateLimit.Reset.Time
		}

		stats[i] = stat
		tokenInfo.mu.RUnlock()
	}

	return stats
}

// RefreshAllTokens refreshes rate limit info for all tokens
func (p *TokenPool) RefreshAllTokens(ctx context.Context) {
	p.mu.RLock()
	tokens := p.tokens
	p.mu.RUnlock()

	for i, tokenInfo := range tokens {
		err := tokenInfo.UpdateRateLimit(ctx)
		if err != nil {
			log.Printf("Failed to refresh token %d: %v", i, err)
		}
	}
}
