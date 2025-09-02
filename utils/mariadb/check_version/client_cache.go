package check_version

import (
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

// createHTTPClient creates a configured Resty client
func createHTTPClient(timeout time.Duration) *resty.Client {
	return resty.New().
		SetTimeout(timeout).
		SetHeader("User-Agent", DefaultUserAgent)
}

// package-level HTTP client reused to benefit from keep-alive
var (
	httpClient *resty.Client
	clientOnce sync.Once
)

func getHTTPClient(timeout time.Duration) *resty.Client {
	clientOnce.Do(func() {
		httpClient = createHTTPClient(timeout)
	})
	// update timeout for each call
	httpClient.SetTimeout(timeout)
	return httpClient
}

// simple in-memory cache to avoid repeated external calls for same major.minor
type cacheEntry struct {
	value   string
	expires time.Time
}

var (
	eolCacheMu sync.RWMutex
	eolCache   = map[string]cacheEntry{}
	// cache TTL for external EOL queries
	eolCacheTTL = 1 * time.Hour
)

// normalizeVersion cleans common prefixes/suffixes and returns major, minor
// returns ok=false when unable to produce major/minor
func normalizeVersion(version string) (major, minor string, ok bool) {
	if version == "" {
		return "", "", false
	}
	v := strings.ToLower(strings.TrimSpace(version))
	// drop build metadata and pre-release suffixes like -rc, -beta, +meta
	if idx := strings.IndexAny(v, "-+"); idx >= 0 {
		v = v[:idx]
	}
	v = strings.TrimPrefix(v, "v")
	parts := strings.Split(v, ".")
	if len(parts) < 2 {
		return "", "", false
	}
	// keep only numeric portion of major/minor
	major = parts[0]
	minor = parts[1]
	// strip any non-digit chars (defensive)
	// note: regexp used originally is heavier; simple loop is lighter here
	strip := func(s string) string {
		out := make([]byte, 0, len(s))
		for i := 0; i < len(s); i++ {
			c := s[i]
			if c >= '0' && c <= '9' {
				out = append(out, c)
			}
		}
		return string(out)
	}
	major = strip(major)
	minor = strip(minor)
	if major == "" || minor == "" {
		return "", "", false
	}
	return major, minor, true
}
