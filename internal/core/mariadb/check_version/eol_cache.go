package check_version

import (
	"sync"
	"time"

	"sfDBTools/utils/mariadb"
)

type eolEntry struct {
	value   string
	expires time.Time
}

var (
	eolMu    sync.RWMutex
	eolCache = map[string]eolEntry{}
	// TTL for cached EOL values
	eolTTL = 12 * time.Hour
)

// GetEOLCached returns EOL date for version, caching the result for eolTTL.
// It calls mariadb.GetMariaDBEOLDate only on cache miss.
func GetEOLCached(version string) string {
	if version == "" {
		return ""
	}

	// use version string directly as cache key to avoid extra deps
	majorMinor := version

	eolMu.RLock()
	if ent, ok := eolCache[majorMinor]; ok {
		if time.Now().Before(ent.expires) {
			eolMu.RUnlock()
			return ent.value
		}
	}
	eolMu.RUnlock()

	// miss: compute
	val := mariadb.GetMariaDBEOLDate(version)

	eolMu.Lock()
	eolCache[majorMinor] = eolEntry{value: val, expires: time.Now().Add(eolTTL)}
	eolMu.Unlock()
	return val
}
