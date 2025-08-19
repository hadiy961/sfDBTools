package common

import (
	"strings"

	"sfDBTools/internal/logger"
	"sfDBTools/utils/database"
)

// ParseGranteeFormat parses grantee format 'user'@'host'
func ParseGranteeFormat(grantee string, lg *logger.Logger) (string, string, bool) {
	grantee = strings.Trim(grantee, "'")
	parts := strings.Split(grantee, "'@'")
	if len(parts) != 2 {
		lg.Warn("Invalid grantee format", logger.String("grantee", grantee))
		return "", "", false
	}
	return parts[0], parts[1], true
}

// ShouldSkipUser determines if a user should be skipped
func ShouldSkipUser(username string, includeSystemUser bool) bool {
	return !includeSystemUser && database.IsSystemUser(username)
}
