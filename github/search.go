package github

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/go-github/v57/github"
)

// SearchOptions represents search options
type SearchOptions struct {
	Keywords    []string
	MatchType   string   // "precise" or "fuzzy"
	ExcludeExts []string
	Language    string
	Sort        string // "indexed", "stars", "forks", etc.
	Order       string // "asc" or "desc"
}

// SearchResultItem represents a single search result
type SearchResultItem struct {
	RepoFullName    string    `json:"repo_full_name"`
	RepoURL         string    `json:"repo_url"`
	FilePath        string    `json:"file_path"`
	FileURL         string    `json:"file_url"`
	HTMLURL         string    `json:"html_url"`
	MatchedKeywords []string  `json:"matched_keywords"`
	ContentSnippet  string    `json:"content_snippet"`
	Score           float64   `json:"score"`
	CreatedAt       time.Time `json:"created_at"`
}

// SearchService handles GitHub code search
type SearchService struct {
	tokenPool *TokenPool
}

// NewSearchService creates a new search service
func NewSearchService(tokenPool *TokenPool) *SearchService {
	return &SearchService{
		tokenPool: tokenPool,
	}
}

// SearchCode performs a GitHub code search
func (s *SearchService) SearchCode(ctx context.Context, opts SearchOptions) ([]*SearchResultItem, error) {
	query := s.buildQuery(opts)
	log.Printf("Executing search query: %s", query)

	client, tokenInfo, err := s.tokenPool.GetClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %w", err)
	}

	searchOpts := &github.SearchOptions{
		Sort:  opts.Sort,
		Order: opts.Order,
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	}

	results := make([]*SearchResultItem, 0)
	page := 1

	for {
		searchOpts.Page = page

		// Perform search
		codeResults, resp, err := client.Search.Code(ctx, query, searchOpts)
		if err != nil {
			// Check if it's a rate limit error
			if resp != nil && resp.StatusCode == 403 {
				log.Printf("Rate limit hit, token stats: %+v", tokenInfo)
				return nil, fmt.Errorf("rate limit exceeded: %w", err)
			}
			return nil, fmt.Errorf("search failed: %w", err)
		}

		// Process results
		for _, result := range codeResults.CodeResults {
			item := s.convertToSearchResultItem(result, opts.Keywords)
			if item != nil {
				results = append(results, item)
			}
		}

		log.Printf("Page %d: Found %d results, Total: %d", page, len(codeResults.CodeResults), codeResults.GetTotal())

		// Check if there are more pages
		if page >= 10 || len(codeResults.CodeResults) == 0 {
			// GitHub API limits to 1000 results (10 pages * 100 per page)
			break
		}

		page++

		// Rate limiting: wait between requests
		time.Sleep(2 * time.Second)
	}

	log.Printf("Search completed: %d total results", len(results))
	return results, nil
}

// buildQuery builds a GitHub search query from options
func (s *SearchService) buildQuery(opts SearchOptions) string {
	var queryParts []string

	if opts.MatchType == "precise" {
		// Precise match: use quotes for exact phrase matching
		for _, keyword := range opts.Keywords {
			if keyword != "" {
				queryParts = append(queryParts, fmt.Sprintf(`"%s"`, keyword))
			}
		}
	} else {
		// Fuzzy match: combine keywords
		for _, keyword := range opts.Keywords {
			if keyword != "" {
				queryParts = append(queryParts, keyword)
			}
		}
	}

	query := strings.Join(queryParts, " ")

	// Exclude file extensions
	for _, ext := range opts.ExcludeExts {
		if ext != "" {
			query += fmt.Sprintf(" -extension:%s", strings.TrimPrefix(ext, "."))
		}
	}

	// Add language filter if specified
	if opts.Language != "" {
		query += fmt.Sprintf(" language:%s", opts.Language)
	}

	return query
}

// convertToSearchResultItem converts a GitHub code result to our format
func (s *SearchService) convertToSearchResultItem(result *github.CodeResult, keywords []string) *SearchResultItem {
	if result == nil || result.Repository == nil {
		return nil
	}

	item := &SearchResultItem{
		RepoFullName:    result.Repository.GetFullName(),
		RepoURL:         result.Repository.GetHTMLURL(),
		FilePath:        result.GetPath(),
		FileURL:         result.GetHTMLURL(), // Use HTML URL as file URL
		HTMLURL:         result.GetHTMLURL(),
		MatchedKeywords: s.findMatchedKeywords(result, keywords),
		ContentSnippet:  s.extractSnippet(result),
		Score:           1.0, // Default score, can be enhanced later
		CreatedAt:       time.Now(),
	}

	return item
}

// findMatchedKeywords finds which keywords were matched in the result
func (s *SearchService) findMatchedKeywords(result *github.CodeResult, keywords []string) []string {
	matched := make([]string, 0)
	content := strings.ToLower(result.GetName() + " " + result.GetPath())

	if result.TextMatches != nil {
		for _, match := range result.TextMatches {
			content += " " + strings.ToLower(match.GetFragment())
		}
	}

	for _, keyword := range keywords {
		if keyword != "" && strings.Contains(content, strings.ToLower(keyword)) {
			matched = append(matched, keyword)
		}
	}

	return matched
}

// extractSnippet extracts a content snippet from the search result
func (s *SearchService) extractSnippet(result *github.CodeResult) string {
	if result.TextMatches != nil && len(result.TextMatches) > 0 {
		// Use the first text match as snippet
		snippet := result.TextMatches[0].GetFragment()
		if len(snippet) > 500 {
			snippet = snippet[:500] + "..."
		}
		return snippet
	}

	return ""
}

// SearchWithRetry performs a search with automatic retry on rate limit
func (s *SearchService) SearchWithRetry(ctx context.Context, opts SearchOptions, maxRetries int) ([]*SearchResultItem, error) {
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		results, err := s.SearchCode(ctx, opts)
		if err == nil {
			return results, nil
		}

		lastErr = err

		if strings.Contains(err.Error(), "rate limit") {
			log.Printf("Rate limit hit, attempt %d/%d, waiting before retry...", i+1, maxRetries)
			time.Sleep(time.Duration(i+1) * 10 * time.Second)
			continue
		}

		// For other errors, don't retry
		return nil, err
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// ParseKeywords parses keywords from JSON string
func ParseKeywords(keywordsJSON string) ([]string, error) {
	var keywords []string
	err := json.Unmarshal([]byte(keywordsJSON), &keywords)
	if err != nil {
		return nil, err
	}
	return keywords, nil
}

// ParseExcludeExts parses exclude extensions from JSON string
func ParseExcludeExts(extsJSON string) ([]string, error) {
	if extsJSON == "" {
		return []string{}, nil
	}

	var exts []string
	err := json.Unmarshal([]byte(extsJSON), &exts)
	if err != nil {
		return nil, err
	}
	return exts, nil
}
