package template

import (
	"fmt"
	"strings"
)

func (t *MariaDBConfigTemplate) GenerateConfigFromTemplate(values map[string]string) (string, error) {
	if t.Content == "" {
		return "", fmt.Errorf("template content is empty")
	}
	result := t.Content
	for key, placeholder := range t.Placeholders {
		placeholderPattern := fmt.Sprintf("{{%s}}", placeholder)
		var value string
		if val, exists := values[key]; exists {
			value = val
		} else if val, exists := t.DefaultValues[key]; exists {
			value = val
		} else {
			return "", fmt.Errorf("no value provided for placeholder %s (key: %s)", placeholder, key)
		}
		result = strings.ReplaceAll(result, placeholderPattern, value)
	}
	return result, nil
}
