package common

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ParseArgsString parses a string of command-line arguments into a slice
func ParseArgsString(argsStr string) []string {
	var args []string
	var current []rune
	inQuotes := false
	for _, char := range argsStr {
		switch char {
		case '"':
			inQuotes = !inQuotes
		case ' ':
			if inQuotes {
				current = append(current, char)
			} else if len(current) > 0 {
				args = append(args, string(current))
				current = nil
			}
		default:
			current = append(current, char)
		}
	}
	if len(current) > 0 {
		args = append(args, string(current))
	}
	return args
}

func EscapeValue(val interface{}) string {
	switch v := val.(type) {
	case nil:
		return "NULL"
	case []byte:
		return fmt.Sprintf("0x%x", v)
	case string:
		return "'" + strings.ReplaceAll(v, "'", "''") + "'"
	default:
		return fmt.Sprintf("'%v'", v)
	}
}

// RemoveDataFlags removes any data-related flags from args
func RemoveDataFlags(args []string) []string {
	skip := map[string]struct{}{
		"--no-data": {}, "--no-data=true": {}, "--no-data=false": {},
	}
	var filtered []string
	for _, arg := range args {
		if _, found := skip[arg]; !found {
			filtered = append(filtered, arg)
		}
	}
	return filtered
}

// ReadDatabaseList reads database names from a text file
func ReadDatabaseList(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	var databases []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			databases = append(databases, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", filePath, err)
	}

	return databases, nil
}
