// Package crawler provides website crawling with SPA detection.
package crawler

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	vibium "github.com/agentplexus/vibium-go"
)

// Crawler crawls websites discovering pages for accessibility auditing.
type Crawler struct {
	vibe   *vibium.Vibe
	logger *slog.Logger
	config Config
}

// Config configures the crawler.
type Config struct {
	// Maximum crawl depth
	MaxDepth int

	// Maximum pages to crawl
	MaxPages int

	// Include patterns (glob)
	IncludePatterns []string

	// Exclude patterns (glob)
	ExcludePatterns []string

	// Wait for SPA hydration
	WaitForSPA bool

	// SPA framework indicators
	SPAIndicators []string

	// Respect robots.txt
	RespectRobots bool

	// Use sitemap.xml
	UseSitemap bool

	// Delay between requests
	Delay time.Duration

	// Request timeout
	Timeout time.Duration

	// Concurrency limit
	Concurrency int
}

// DefaultConfig returns default crawler configuration.
func DefaultConfig() Config {
	return Config{
		MaxDepth:    3,
		MaxPages:    100,
		WaitForSPA:  true,
		SPAIndicators: []string{
			"[data-reactroot]",
			"#__next",
			"[data-v-]",
			"#app[data-server-rendered]",
			"[ng-app]",
			"[ng-controller]",
		},
		RespectRobots: true,
		UseSitemap:    true,
		Delay:         500 * time.Millisecond,
		Timeout:       30 * time.Second,
		Concurrency:   3,
	}
}

// Page represents a discovered page.
type Page struct {
	URL         string        `json:"url"`
	Title       string        `json:"title"`
	Depth       int           `json:"depth"`
	DiscoveredFrom string     `json:"discoveredFrom"`
	IsSPA       bool          `json:"isSPA"`
	SPAFramework string       `json:"spaFramework,omitempty"`
	LoadTime    time.Duration `json:"loadTime"`
	Links       []string      `json:"links"`
	Error       string        `json:"error,omitempty"`
}

// Result contains the crawl results.
type Result struct {
	StartURL    string        `json:"startUrl"`
	Pages       []Page        `json:"pages"`
	TotalPages  int           `json:"totalPages"`
	Duration    time.Duration `json:"duration"`
	RobotsTxt   string        `json:"robotsTxt,omitempty"`
	SitemapURLs []string      `json:"sitemapUrls,omitempty"`
}

// NewCrawler creates a new crawler.
func NewCrawler(vibe *vibium.Vibe, logger *slog.Logger, config Config) *Crawler {
	return &Crawler{
		vibe:   vibe,
		logger: logger,
		config: config,
	}
}

// Crawl crawls a website starting from the given URL.
func (c *Crawler) Crawl(ctx context.Context, startURL string) (*Result, error) {
	startTime := time.Now()

	parsedStart, err := url.Parse(startURL)
	if err != nil {
		return nil, fmt.Errorf("invalid start URL: %w", err)
	}

	result := &Result{
		StartURL: startURL,
		Pages:    make([]Page, 0),
	}

	// Check robots.txt
	if c.config.RespectRobots {
		robotsURL := fmt.Sprintf("%s://%s/robots.txt", parsedStart.Scheme, parsedStart.Host)
		robotsTxt, err := c.fetchRobotsTxt(ctx, robotsURL)
		if err == nil {
			result.RobotsTxt = robotsTxt
		}
	}

	// Get sitemap URLs
	if c.config.UseSitemap {
		sitemapURLs := c.fetchSitemapURLs(ctx, parsedStart)
		result.SitemapURLs = sitemapURLs
	}

	// BFS crawl
	visited := make(map[string]bool)
	queue := []queueItem{{url: startURL, depth: 0, from: ""}}
	var mu sync.Mutex

	for len(queue) > 0 && len(result.Pages) < c.config.MaxPages {
		select {
		case <-ctx.Done():
			result.Duration = time.Since(startTime)
			result.TotalPages = len(result.Pages)
			return result, ctx.Err()
		default:
		}

		// Get next URL
		mu.Lock()
		if len(queue) == 0 {
			mu.Unlock()
			break
		}
		item := queue[0]
		queue = queue[1:]

		// Skip if already visited
		normalizedURL := c.normalizeURL(item.url)
		if visited[normalizedURL] {
			mu.Unlock()
			continue
		}
		visited[normalizedURL] = true
		mu.Unlock()

		// Check depth limit
		if item.depth > c.config.MaxDepth {
			continue
		}

		// Check patterns
		if !c.matchesPatterns(item.url) {
			continue
		}

		// Delay between requests
		if c.config.Delay > 0 {
			time.Sleep(c.config.Delay)
		}

		// Crawl page
		page, err := c.crawlPage(ctx, item.url, item.depth, item.from)
		if err != nil {
			c.logger.Warn("failed to crawl page", "url", item.url, "error", err)
			page = &Page{
				URL:           item.url,
				Depth:         item.depth,
				DiscoveredFrom: item.from,
				Error:         err.Error(),
			}
		}

		mu.Lock()
		result.Pages = append(result.Pages, *page)

		// Add discovered links to queue
		if page.Error == "" {
			for _, link := range page.Links {
				linkParsed, err := url.Parse(link)
				if err != nil {
					continue
				}

				// Only crawl same domain
				if linkParsed.Host != "" && linkParsed.Host != parsedStart.Host {
					continue
				}

				// Resolve relative URLs
				resolvedURL := parsedStart.ResolveReference(linkParsed).String()
				normalizedLink := c.normalizeURL(resolvedURL)

				if !visited[normalizedLink] && len(result.Pages)+len(queue) < c.config.MaxPages {
					queue = append(queue, queueItem{
						url:   resolvedURL,
						depth: item.depth + 1,
						from:  item.url,
					})
				}
			}
		}
		mu.Unlock()
	}

	result.Duration = time.Since(startTime)
	result.TotalPages = len(result.Pages)
	return result, nil
}

type queueItem struct {
	url   string
	depth int
	from  string
}

// crawlPage crawls a single page and extracts information.
func (c *Crawler) crawlPage(ctx context.Context, pageURL string, depth int, from string) (*Page, error) {
	startTime := time.Now()

	page := &Page{
		URL:           pageURL,
		Depth:         depth,
		DiscoveredFrom: from,
	}

	// Navigate to page
	if err := c.vibe.Go(ctx, pageURL); err != nil {
		return nil, fmt.Errorf("navigation failed: %w", err)
	}

	// Wait for page load
	if err := c.waitForPageLoad(ctx); err != nil {
		c.logger.Warn("page load wait failed", "url", pageURL, "error", err)
	}

	// Detect SPA
	spaFramework := c.detectSPAFramework(ctx)
	if spaFramework != "" {
		page.IsSPA = true
		page.SPAFramework = spaFramework

		// Wait for SPA hydration
		if c.config.WaitForSPA {
			if err := c.waitForSPAHydration(ctx, spaFramework); err != nil {
				c.logger.Warn("SPA hydration wait failed", "url", pageURL, "error", err)
			}
		}
	}

	page.LoadTime = time.Since(startTime)

	// Get page title
	title, err := c.vibe.Title(ctx)
	if err == nil {
		page.Title = title
	}

	// Extract links
	links, err := c.extractLinks(ctx)
	if err != nil {
		c.logger.Warn("failed to extract links", "url", pageURL, "error", err)
	} else {
		page.Links = links
	}

	return page, nil
}

// detectSPAFramework detects which SPA framework is used.
func (c *Crawler) detectSPAFramework(ctx context.Context) string {
	frameworks := map[string]string{
		"[data-reactroot]":           "react",
		"#__next":                    "next.js",
		"[data-v-]":                  "vue",
		"#app[data-server-rendered]": "nuxt",
		"[ng-app]":                   "angular",
		"[ng-controller]":            "angular",
		"[data-ember-action]":        "ember",
		"[data-svelte]":              "svelte",
	}

	for selector, framework := range frameworks {
		el, err := c.vibe.Find(ctx, selector, nil)
		if err == nil && el != nil {
			return framework
		}
	}

	return ""
}

// waitForSPAHydration waits for SPA framework to hydrate.
func (c *Crawler) waitForSPAHydration(ctx context.Context, framework string) error {
	// Framework-specific hydration indicators
	var indicator string
	switch framework {
	case "react", "next.js":
		// Wait for hydration marker or network idle
		indicator = "[data-reactroot]:not([data-react-helmet])"
	case "vue", "nuxt":
		indicator = "[data-v-]:not([data-server-rendered])"
	case "angular":
		indicator = "[ng-scope]"
	default:
		// Generic wait
		time.Sleep(1 * time.Second)
		return nil
	}

	// Use Find with timeout for waiting
	_, err := c.vibe.Find(ctx, indicator, &vibium.FindOptions{
		Timeout: 5 * time.Second,
	})
	return err
}

// waitForPageLoad waits for the page to finish loading.
func (c *Crawler) waitForPageLoad(ctx context.Context) error {
	// Wait for load state
	return c.vibe.WaitForLoad(ctx, "networkidle", c.config.Timeout)
}

// extractLinks extracts all links from the current page.
func (c *Crawler) extractLinks(ctx context.Context) ([]string, error) {
	script := `
		return Array.from(document.querySelectorAll('a[href]'))
			.map(a => a.href)
			.filter(href => href.startsWith('http'))
	`

	result, err := c.vibe.Evaluate(ctx, script)
	if err != nil {
		return nil, err
	}

	// Parse result
	links := make([]string, 0)
	if arr, ok := result.([]any); ok {
		for _, item := range arr {
			if link, ok := item.(string); ok {
				links = append(links, link)
			}
		}
	}

	return links, nil
}

// normalizeURL normalizes a URL for comparison.
func (c *Crawler) normalizeURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// Remove fragment
	parsed.Fragment = ""

	// Remove trailing slash
	parsed.Path = strings.TrimSuffix(parsed.Path, "/")

	// Normalize scheme
	parsed.Scheme = strings.ToLower(parsed.Scheme)
	parsed.Host = strings.ToLower(parsed.Host)

	return parsed.String()
}

// matchesPatterns checks if URL matches include/exclude patterns.
func (c *Crawler) matchesPatterns(rawURL string) bool {
	// Check exclude patterns first
	for _, pattern := range c.config.ExcludePatterns {
		if c.matchGlob(rawURL, pattern) {
			return false
		}
	}

	// If no include patterns, include all
	if len(c.config.IncludePatterns) == 0 {
		return true
	}

	// Check include patterns
	for _, pattern := range c.config.IncludePatterns {
		if c.matchGlob(rawURL, pattern) {
			return true
		}
	}

	return false
}

// matchGlob performs simple glob matching.
func (c *Crawler) matchGlob(s, pattern string) bool {
	// Convert glob to regex
	pattern = regexp.QuoteMeta(pattern)
	pattern = strings.ReplaceAll(pattern, `\*\*`, ".*")
	pattern = strings.ReplaceAll(pattern, `\*`, "[^/]*")
	pattern = "^" + pattern + "$"

	matched, _ := regexp.MatchString(pattern, s)
	return matched
}

// fetchRobotsTxt fetches robots.txt content.
func (c *Crawler) fetchRobotsTxt(ctx context.Context, robotsURL string) (string, error) {
	// Navigate to robots.txt
	if err := c.vibe.Go(ctx, robotsURL); err != nil {
		return "", err
	}

	// Get content
	content, err := c.vibe.Content(ctx)
	if err != nil {
		return "", err
	}

	return content, nil
}

// fetchSitemapURLs fetches URLs from sitemap.xml.
func (c *Crawler) fetchSitemapURLs(ctx context.Context, baseURL *url.URL) []string {
	sitemapURL := fmt.Sprintf("%s://%s/sitemap.xml", baseURL.Scheme, baseURL.Host)

	// Navigate to sitemap
	if err := c.vibe.Go(ctx, sitemapURL); err != nil {
		return nil
	}

	// Extract URLs
	script := `
		return Array.from(document.querySelectorAll('loc'))
			.map(loc => loc.textContent)
			.filter(url => url && url.startsWith('http'))
	`

	result, err := c.vibe.Evaluate(ctx, script)
	if err != nil {
		return nil
	}

	urls := make([]string, 0)
	if arr, ok := result.([]any); ok {
		for _, item := range arr {
			if u, ok := item.(string); ok {
				urls = append(urls, u)
			}
		}
	}

	return urls
}

// GetSPAFramework returns the detected SPA framework for a page.
func (c *Crawler) GetSPAFramework(ctx context.Context) string {
	return c.detectSPAFramework(ctx)
}

// IsSPAPage checks if the current page is an SPA.
func (c *Crawler) IsSPAPage(ctx context.Context) bool {
	return c.detectSPAFramework(ctx) != ""
}
