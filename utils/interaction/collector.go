package interaction

import (
	"strings"

	"sfDBTools/utils/terminal"
)

// DefaultsResolver provides methods used by the collector to resolve defaults.
type DefaultsResolver interface {
	GetStringDefault(templateKey string, hardcodedDefault string) string
	GetIntDefault(templateKey string, hardcodedDefault int) int
	GetDirectoryFromTemplate(templateKey string) string
	GetBoolDefault(templateKey string, hardcodedDefault bool) bool
}

// InputCollector is a reusable collector for interactive prompts.
type InputCollector struct {
	Defaults DefaultsResolver
}

// NewInputCollector creates a new InputCollector using provided DefaultsResolver.
func NewInputCollector(defaults DefaultsResolver) *InputCollector {
	return &InputCollector{Defaults: defaults}
}

// CollectString collects string input with validation
func (ic *InputCollector) CollectString(question string, currentValue string, templateKey string, hardcodedDefault string, validator func(string) error) (string, error) {
	defaultValue := ic.Defaults.GetStringDefault(templateKey, hardcodedDefault)
	if currentValue != "" {
		defaultValue = currentValue
	}

	input := terminal.AskString(question, defaultValue)
	result := strings.TrimSpace(input)

	if result == "" {
		result = defaultValue
	}

	if validator != nil {
		if err := validator(result); err != nil {
			return "", err
		}
	}

	return result, nil
}

// CollectInt collects integer input with validation
func (ic *InputCollector) CollectInt(question string, currentValue int, templateKey string, validator func(int) error) (int, error) {
	defaultValue := ic.Defaults.GetIntDefault(templateKey, 0)
	if currentValue > 0 {
		defaultValue = currentValue
	}

	result := terminal.AskInt(question, defaultValue)

	if validator != nil {
		if err := validator(result); err != nil {
			return 0, err
		}
	}

	return result, nil
}

// CollectBool collects boolean input (yes/no)
func (ic *InputCollector) CollectBool(question string, defaultValue bool) bool {
	return terminal.AskYesNo(question, defaultValue)
}

// CollectDirectory collects directory using template extraction if available
func (ic *InputCollector) CollectDirectory(question string, currentValue string, templateKey string, hardcodedDefault string, validator func(string) error) (string, error) {
	// Try extract directory from template value if available
	templateDir := ic.Defaults.GetDirectoryFromTemplate(templateKey)

	if templateDir != "" {
		hardcodedDefault = templateDir
	}

	if currentValue != "" {
		hardcodedDefault = currentValue
	}

	return ic.CollectString(question, "", "", hardcodedDefault, validator)
}
