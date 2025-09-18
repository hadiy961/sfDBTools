package interactive

import (
	"strings"

	"sfDBTools/utils/terminal"
)

// InputCollector menyediakan abstraksi untuk mengumpulkan input dengan validasi
// Task 2: Menghilangkan duplikasi code dalam input collection
type InputCollector struct {
	Defaults *ConfigDefaults
}

// NewInputCollector membuat instance baru InputCollector
func NewInputCollector(defaults *ConfigDefaults) *InputCollector {
	return &InputCollector{
		Defaults: defaults,
	}
}

// CollectString mengumpulkan input string dengan validasi
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

// CollectInt mengumpulkan input integer dengan validasi
func (ic *InputCollector) CollectInt(question string, currentValue int, templateKey string, validator func(int) error) (int, error) {
	defaultValue := ic.Defaults.GetIntDefault(templateKey, 0)
	if currentValue > 0 {
		defaultValue = currentValue
	}

	// Use AskInt helper which handles prompting and validation loop
	result := terminal.AskInt(question, defaultValue)

	if validator != nil {
		if err := validator(result); err != nil {
			return 0, err
		}
	}

	return result, nil
}

// CollectBool mengumpulkan input boolean (yes/no)
func (ic *InputCollector) CollectBool(question string, defaultValue bool) bool {
	return terminal.AskYesNo(question, defaultValue)
}

// CollectDirectory khusus untuk directory dengan path extraction dari template
func (ic *InputCollector) CollectDirectory(question string, currentValue string, templateKey string, hardcodedDefault string) (string, error) {
	// Coba extract directory dari template value jika ada
	templateDir := ic.Defaults.GetDirectoryFromTemplate(templateKey)

	if templateDir != "" {
		hardcodedDefault = templateDir
	}

	// Tentukan default value dengan priority
	if currentValue != "" {
		hardcodedDefault = currentValue
	}

	return ic.CollectString(question, "", "", hardcodedDefault, ValidateAbsolutePath)
}
