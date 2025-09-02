package check_version

import (
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

// FetchVersions implements VersionFetcher interface
func (f *HTTPVersionFetcher) FetchVersions() ([]VersionInfo, error) {
	client := &http.Client{Timeout: f.Timeout}

	req, err := http.NewRequest("GET", f.URL, nil)
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

	// Read response body safely
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
