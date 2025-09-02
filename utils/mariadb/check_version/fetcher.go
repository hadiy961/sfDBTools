package check_version

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// NewHTTPVersionFetcher creates a new HTTP-based version fetcher
func NewHTTPVersionFetcher(url string, parser VersionParser) *HTTPVersionFetcher {
	return &HTTPVersionFetcher{
		URL:       url,
		Timeout:   DefaultHTTPTimeout,
		UserAgent: DefaultUserAgent,
		Parser:    parser,
	}
}

// FetchVersions implements VersionFetcher interface (backwards compatible)
func (f *HTTPVersionFetcher) FetchVersions() ([]VersionInfo, error) {
	// Default to background context if caller doesn't supply one
	return f.FetchVersionsWithCtx(context.Background())
}

// FetchVersionsWithCtx fetches versions using provided context for cancellation/timeouts
func (f *HTTPVersionFetcher) FetchVersionsWithCtx(ctx context.Context) ([]VersionInfo, error) {
	// create request with context so caller can cancel
	client := &http.Client{Timeout: f.Timeout}

	req, err := http.NewRequestWithContext(ctx, "GET", f.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", f.UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from %s: %w", f.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, f.URL)
	}

	// Read response body safely with a small deadline derived from context (if any)
	// Respect existing timeout on client as well
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return f.Parser.ParseVersions(string(body))
}

// GetName implements VersionFetcher interface
func (f *HTTPVersionFetcher) GetName() string {
	return fmt.Sprintf("HTTP Fetcher (%s)", f.URL)
}
